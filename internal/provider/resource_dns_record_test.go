package provider

import (
	"context"
	"errors"
	"fmt"
	"testing"

	pihole "github.com/awaybreaktoday/lib-pihole-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccLocalDNS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLocalDNSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testLocalDNSResourceConfig("foo", "foo.com", "127.0.0.1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_dns_record.foo", "domain", "foo.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.foo", "ip", "127.0.0.1"),
					resource.TestCheckNoResourceAttr("pihole_dns_record.foo", "ttl"),
					resource.TestCheckResourceAttr("pihole_dns_record.foo", "comment", ""),
					testCheckLocalDNSResourceExists(t, "foo.com", "127.0.0.1"),
				),
			},
			{
				Config: testLocalDNSResourceConfig("foo", "foo.com", "127.0.0.2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_dns_record.foo", "domain", "foo.com"),
					resource.TestCheckResourceAttr("pihole_dns_record.foo", "ip", "127.0.0.2"),
					resource.TestCheckNoResourceAttr("pihole_dns_record.foo", "ttl"),
					resource.TestCheckResourceAttr("pihole_dns_record.foo", "comment", ""),
					testCheckLocalDNSResourceExists(t, "foo.com", "127.0.0.2"),
				),
			},
		},
	})
}

func testLocalDNSResourceConfig(name string, domain string, ip string) string {
	return fmt.Sprintf(`
		resource "pihole_dns_record" %q {
			domain = %q
			ip     = %q
		}	
	`, name, domain, ip)
}

func testCheckLocalDNSResourceExists(_ *testing.T, domain string, ip string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		client := testAccProvider.Meta().(*pihole.Client)

		record, err := client.LocalDNS.Get(context.Background(), domain)
		if err != nil {
			return err
		}

		if record.IP != ip {
			return fmt.Errorf("requested %s:%s does not match IP: %s", domain, ip, record.IP)
		}

		return nil
	}
}

func testAccCheckLocalDNSDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pihole.Client)

	for _, r := range s.RootModule().Resources {
		if r.Type != "pihole_dns_record" {
			continue
		}

		if _, err := client.LocalDNS.Get(context.Background(), r.Primary.ID); err != nil {
			if !errors.Is(err, pihole.ErrorLocalDNSNotFound) {
				return err
			}
		}
	}

	return nil
}
