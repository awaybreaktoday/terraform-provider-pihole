package provider

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"

	pihole "github.com/awaybreaktoday/lib-pihole-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceCNAMERecords returns a schema resource for listing Pi-hole CNAME records
func dataSourceCNAMERecords() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCNAMERecordsRead,
		Schema: map[string]*schema.Schema{
			"records": {
				Description: "List of CNAME Pi-hole records",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Description: "CNAME record domain",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"target": {
							Description: "CNAME target value where traffic is routed to from the domain",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"ttl": {
							Description: "TTL (in seconds) returned by Pi-hole for the CNAME record.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// dataSourceCNAMERecordsRead lists all Pi-hole CNAME records
func dataSourceCNAMERecordsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	client, ok := meta.(*pihole.Client)
	if !ok {
		return diag.Errorf("Could not load client in resource request")
	}

	cnameList, err := client.LocalCNAME.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	sort.Slice(cnameList, func(i, j int) bool {
		if cnameList[i].Domain == cnameList[j].Domain {
			return cnameList[i].Target < cnameList[j].Target
		}

		return cnameList[i].Domain < cnameList[j].Domain
	})

	list := make([]map[string]interface{}, len(cnameList))
	hash := sha256.New()

	for i, r := range cnameList {
		hash.Write([]byte(r.Domain))
		hash.Write([]byte{0})
		hash.Write([]byte(r.Target))
		hash.Write([]byte{0})
		hash.Write([]byte(strconv.Itoa(r.TTL)))
		hash.Write([]byte{0})

		list[i] = map[string]interface{}{
			"domain": r.Domain,
			"target": r.Target,
			"ttl":    r.TTL,
		}
	}

	if err := d.Set("records", list); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%x", hash.Sum(nil)))

	return diags
}
