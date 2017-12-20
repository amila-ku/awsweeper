package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws"
	"github.com/spf13/afero"
)

func TestVpc_tags(t *testing.T) {
	argsDryRun := []string{"cmd", "--profile=learning", "--force", "config.yml"}
	args := []string{"cmd", "--profile=learning", "--force", "config.yml"}

	osFs = afero.NewMemMapFs()
	afero.WriteFile(osFs, "config.yml", []byte(bla), 0644)

	var vpc ec2.Vpc
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc),
					testMain(argsDryRun),
					testVpcExists(&vpc),
					testMain(args),
					testVpcDeleted(&vpc),
				),
			},
		},
	})
}

const bla = `
aws_vpc:
  tags:
    foo: bar
`

func testMain(args []string) resource.TestCheckFunc {
	os.Args = args

	return func(s *terraform.State) error {
		main()
		return nil
	}
}

const testAccVpcConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"

	tags {
		foo = "bar"
	}
}
`

func testAccCheckVpcDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc" {
			continue
		}

		// Try to find the VPC
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err == nil {
			if len(resp.Vpcs) > 0 {
				return fmt.Errorf("VPCs still exist.")
			}

			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidVpcID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckVpcExists(n string, vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC ID is set")
		}

		conn := testAccProvider.Meta().(*aws/AWSClient).ec2conn
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err != nil {
			return err
		}
		if len(resp.Vpcs) == 0 {
			return fmt.Errorf("VPC not found")
		}

		*vpc = *resp.Vpcs[0]

		return nil
	}
}

func testVpcDeleted(vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{vpc.VpcId},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err != nil {
			return err
		}
		if len(resp.Vpcs) != 0 {
			return fmt.Errorf("VPC hasn't been deleted")
		}

		return nil
	}
}

func testVpcExists(vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{vpc.VpcId},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err != nil {
			return err
		}
		if len(resp.Vpcs) == 0 {
			return fmt.Errorf("VPC has been deleted")
		}

		return nil
	}
}
