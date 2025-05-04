package main

import (
	"fmt"

	"github.com/eedygreen/pulumi/infra/vpc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {

	pulumi.Run(func(ctx *pulumi.Context) error {

		Vpc, err := vpc.CreateVpc(ctx, vpc.Parameters{})
		if err != nil {
			return fmt.Errorf("failed to create VPC: %v", err)
		}
		validConfig, err := vpc.Validate(ctx)
		if err != nil {
			return fmt.Errorf("failed to load configurations: %w", err)
		}
		subnets, err := vpc.CreateSubnets(ctx, Vpc, validConfig)
		if err != nil {
			return fmt.Errorf("failed to create subnets: %v", err)
		}

		ctx.Export("VpcId", Vpc.ID())
		ctx.Export("VpcCidr", Vpc.CidrBlock)
		ctx.Export("PublicSubnetIds", subnets[0].ID())
		ctx.Export("PrivateSubnetIds", subnets[1].ID())
		ctx.Export("PublicSubnetAz", subnets[0].AvailabilityZone)
		ctx.Export("PrivateSubnetAz", subnets[1].AvailabilityZone)
		ctx.Export("PublicSubnetCidr", subnets[0].CidrBlock)
		ctx.Export("PrivateSubnetCidr", subnets[1].CidrBlock)

		return nil
	})
}
