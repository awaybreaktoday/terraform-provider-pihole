package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testAccPreCheck(t *testing.T) {
	url := os.Getenv("PIHOLE_URL")
	if url == "" {
		t.Fatal("PIHOLE_URL must be set for acceptance tests")
	}

	password := os.Getenv("PIHOLE_PASSWORD")
	apiToken := os.Getenv("PIHOLE_API_TOKEN")
	if password == "" && apiToken == "" {
		t.Fatal("PIHOLE_PASSWORD or PIHOLE_API_TOKEN must be set for acceptance tests")
	}

	if password != "" && os.Getenv("__PIHOLE_SESSION_ID") == "" {
		t.Log("No session ID found, setting for testing")

		client, err := Config{
			URL:      url,
			Password: password,
			APIToken: apiToken,
		}.Client(context.TODO())

		if err != nil {
			t.Fatal(err.Error())
		}
		session, err := client.SessionAPI.Post(context.TODO())
		if err != nil {
			t.Fatal(err.Error())
		}
		if err := os.Setenv("__PIHOLE_SESSION_ID", session.SID); err != nil {
			t.Fatal(err.Error())
		}
	}
}

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"pihole": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderImpl(t *testing.T) {
	var _ *schema.Provider = Provider()
}
