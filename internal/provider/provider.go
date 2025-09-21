package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/markjoyeuxcom/terraform-provider-pihole/internal/version"
)

func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("PIHOLE_PASSWORD", nil),
				Description:  "The admin password used to login to the admin dashboard.",
				Sensitive:    true,
				ExactlyOneOf: []string{"password", "api_token"},
				AtLeastOneOf: []string{"password", "api_token"},
			},
			"api_token": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("PIHOLE_API_TOKEN", nil),
				Description:  "Pi-hole API token used for token-based authentication.",
				Sensitive:    true,
				ExactlyOneOf: []string{"password", "api_token"},
				AtLeastOneOf: []string{"password", "api_token"},
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PIHOLE_URL", "http://pi.hole"),
				Description: "URL where Pi-hole is deployed",
			},
			"ca_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PIHOLE_CA_FILE", nil),
				Description: "CA file to connect to Pi-hole with TLS",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"pihole_cname_records": dataSourceCNAMERecords(),
			"pihole_dns_records":   dataSourceDNSRecords(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"pihole_cname_record": resourceCNAMERecord(),
			"pihole_dns_record":   resourceDNSRecord(),
		},
	}

	provider.ConfigureContextFunc = configure(version.ProviderVersion, provider)

	return provider
}

// configure configures a Pi-hole client to be used for terraform resource requests
func configure(version string, provider *schema.Provider) func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (client interface{}, diags diag.Diagnostics) {
		client, err := Config{
			Password:  d.Get("password").(string),
			APIToken:  d.Get("api_token").(string),
			URL:       d.Get("url").(string),
			UserAgent: provider.UserAgent("terraform-provider-pihole", version),
			CAFile:    d.Get("ca_file").(string),
			SessionID: os.Getenv("__PIHOLE_SESSION_ID"),
		}.Client(ctx)

		if err != nil {
			return nil, diag.FromErr(fmt.Errorf("failed to instantiate client: %w", err))
		}

		return client, diags
	}
}
