package vpc

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type Parameters struct {
	Name               string
	Cidr               string
	Env                string
	VpcTagKey          string
	PublicSubnetCirdr  string
	PrivateSubnetCirdr string
}

func CreateVpc(ctx *pulumi.Context, name string, opt ...Parameters) (pulumi.IDOutput, error) {

	parameters := Parameters{
		Name:               config.New(ctx, "").Require("vpcName"),
		Cidr:               config.New(ctx, "").Require("vpcCIDR"),
		Env:                ctx.Stack(),
		VpcTagKey:          "Name",
		PublicSubnetCirdr:  config.New(ctx, "").Require("publicSubnetCIDR"),
		PrivateSubnetCirdr: config.New(ctx, "").Require("privateSubnetCIDR"),
	}

	switch {
	case parameters.Name == "":
		return pulumi.IDOutput{}, fmt.Errorf("vpcName is required in pulumi config")

	case parameters.Cidr == "":
		return pulumi.IDOutput{}, fmt.Errorf("vpcCIDR is required in pulumi config")

	case parameters.PublicSubnetCirdr == "":
		return pulumi.IDOutput{}, fmt.Errorf("publicSubnetCIDR is required in pulumi config")

	case parameters.PrivateSubnetCirdr == "":
		return pulumi.IDOutput{}, fmt.Errorf("privateSubnetCIDR is required in pulumi config")
	}

	Azs, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{
		State: pulumi.StringRef("available")}, nil)
	if err != nil {
		return pulumi.IDOutput{}, err
	}

	existingVpcs, err := ec2.GetVpcs(ctx, &ec2.GetVpcsArgs{
		Tags: map[string]string{
			parameters.VpcTagKey: parameters.Name,
		},
	})

	if err != nil {
		return pulumi.IDOutput{}, err
	}

	var vpc *ec2.Vpc

	if len(existingVpcs.Ids) > 0 {

		vpc, err = ec2.NewVpc(ctx, parameters.Name, &ec2.VpcArgs{
			CidrBlock: pulumi.String(parameters.Cidr),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(parameters.Name),
			},
		}, pulumi.Import(pulumi.ID(existingVpcs.Ids[0])))

		if err != nil {
			return pulumi.IDOutput{}, err
		}

		ctx.Log.Info(fmt.Sprintf("VPC already exists with ID: %v", existingVpcs.Ids[0]), nil)

	} else {
		fmt.Println("VPC does not exist, creating a new one...")

		vpc, err := ec2.NewVpc(ctx, parameters.Name, &ec2.VpcArgs{
			CidrBlock: pulumi.String(parameters.Cidr),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(parameters.Name),
			},
		})

		if err != nil {
			return pulumi.IDOutput{}, err
		}
		ctx.Log.Info(fmt.Sprintf("Created VPC with ID: %v", vpc.ID()), nil)

	}

	publicSubnet, err := ec2.NewSubnet(ctx, "publicSubnet", &ec2.SubnetArgs{
		VpcId:            vpc.ID(),
		CidrBlock:        pulumi.String(parameters.PublicSubnetCirdr),
		AvailabilityZone: pulumi.String(Azs.Names[0]),
		Tags: pulumi.StringMap{
			"Name": pulumi.Sprintf("%s-public-subnet-%s", parameters.Name, parameters.Env),
		},
	})

	if err != nil {
		return pulumi.IDOutput{}, err
	}

	privateSubnet, err := ec2.NewSubnet(ctx, "privateSubnet", &ec2.SubnetArgs{
		VpcId:            vpc.ID(),
		CidrBlock:        pulumi.String(parameters.PrivateSubnetCirdr),
		AvailabilityZone: pulumi.String(Azs.Names[1]),

		Tags: pulumi.StringMap{
			"Name": pulumi.Sprintf("%s-public-subnet-%s", parameters.Name, parameters.Env),
		},
	})

	if err != nil {
		return pulumi.IDOutput{}, err
	}

	ctx.Export("VpcId", vpc.ID())
	ctx.Export("PublicSubnetIds", publicSubnet.ID())
	ctx.Export("PrivateSubnetIds", privateSubnet.ID())
	ctx.Export("PublicSubnetAz", publicSubnet.AvailabilityZone)
	ctx.Export("PrivateSubnetAz", privateSubnet.AvailabilityZone)

	return vpc.ID(), nil
}
