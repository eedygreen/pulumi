package main

import (
	"fmt"

	"github.com/eedygreen/pulumi/infra/vpc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {

	pulumi.Run(func(ctx *pulumi.Context) error {

		_, err := vpc.CreateVpc(ctx, "awsVpc", vpc.Parameters{})
		if err != nil {
			return fmt.Errorf("failed to create VPC: %v", err)
		}

		return nil
	})
}
