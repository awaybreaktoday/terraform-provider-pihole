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

// dataSourceDNSRecords returns a schema resource for listing Pi-hole local DNS records
func dataSourceDNSRecords() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDNSRecordsRead,
		Schema: map[string]*schema.Schema{
			"records": {
				Description: "List of Pi-hole DNS records",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Description: "DNS record domain",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"ip": {
							Description: "IP address where traffic is routed to from the DNS record domain",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"ttl": {
							Description: "TTL (in seconds) returned by Pi-hole for the DNS record.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"comment": {
							Description: "Comment associated with the DNS record, if present.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// dataSourceDNSRecordsRead lists all Pi-hole local DNS records
func dataSourceDNSRecordsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	client, ok := meta.(*pihole.Client)
	if !ok {
		return diag.Errorf("Could not load client in resource request")
	}

	dnsList, err := client.LocalDNS.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	sort.Slice(dnsList, func(i, j int) bool {
		if dnsList[i].Domain == dnsList[j].Domain {
			return dnsList[i].IP < dnsList[j].IP
		}

		return dnsList[i].Domain < dnsList[j].Domain
	})

	list := make([]map[string]interface{}, len(dnsList))
	hash := sha256.New()

	for i, r := range dnsList {
		hash.Write([]byte(r.Domain))
		hash.Write([]byte{0})
		hash.Write([]byte(r.IP))
		hash.Write([]byte{0})
		hash.Write([]byte(strconv.Itoa(r.TTL)))
		hash.Write([]byte{0})
		hash.Write([]byte(r.Comment))
		hash.Write([]byte{0})

		list[i] = map[string]interface{}{
			"domain":  r.Domain,
			"ip":      r.IP,
			"ttl":     r.TTL,
			"comment": r.Comment,
		}
	}

	if err := d.Set("records", list); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%x", hash.Sum(nil)))

	return diags
}
