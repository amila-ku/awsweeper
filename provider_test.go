package main

import (
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform/builtin/providers/aws"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvidersWithTLS map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider
var testAccTemplateProvider *schema.Provider

func init() {
	testAccProvider = aws.Provider().(*schema.Provider)
	//testAccTemplateProvider = template.Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"aws": testAccProvider,
		//"template": testAccTemplateProvider,
	}
	testAccProvidersWithTLS = map[string]terraform.ResourceProvider{
	//"tls": tls.Provider(),
	}

	for k, v := range testAccProviders {
		testAccProvidersWithTLS[k] = v
	}
}

func TestProvider(t *testing.T) {
	if err := aws.Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = aws.Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("AWS_PROFILE"); v == "" {
		if v := os.Getenv("AWS_ACCESS_KEY_ID"); v == "" {
			t.Fatal("AWS_ACCESS_KEY_ID must be set for acceptance tests")
		}
		if v := os.Getenv("AWS_SECRET_ACCESS_KEY"); v == "" {
			t.Fatal("AWS_SECRET_ACCESS_KEY must be set for acceptance tests")
		}
	}
	if v := os.Getenv("AWS_DEFAULT_REGION"); v == "" {
		log.Println("[INFO] Test: Using us-west-2 as test region")
		os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	}
	err := testAccProvider.Configure(terraform.NewResourceConfig(nil))
	if err != nil {
		t.Fatal(err)
	}
}
