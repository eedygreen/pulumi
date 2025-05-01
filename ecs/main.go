package main

import (
	"fmt"

	"github.com/eedygreen/pulumi/ecs/vpc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {

	pulumi.Run(func(ctx *pulumi.Context) error {

		_, err := vpc.CreateVpc(ctx, "ecsVpc", vpc.Parameters{})
		if err != nil {
			return fmt.Errorf("failed to create VPC: %v", err)
		}

		return nil
	})
}
