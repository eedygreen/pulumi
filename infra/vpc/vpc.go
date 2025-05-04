package vpc

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type Parameters struct {
	Name, Cidr, Env, VpcTagKey, PublicSubnetCirdr, PrivateSubnetCirdr string
}

func (p *Parameters) Validate() error {
	switch {
	case p.Name == "":
		return fmt.Errorf("vpcName is required in pulumi config")

	case p.Cidr == "":
		return fmt.Errorf("vpcCIDR is required in pulumi config")

	case p.PublicSubnetCirdr == "":
		return fmt.Errorf("publicSubnetCIDR is required in pulumi config")

	case p.PrivateSubnetCirdr == "":
		return fmt.Errorf("privateSubnetCIDR is required in pulumi config")

	}
	return nil
}

func CreateVpc(ctx *pulumi.Context, opt ...Parameters) error {
	conf := config.New(ctx, "")
	parameters := Parameters{
		Name:               conf.Require("vpcName"),
		Cidr:               conf.Require("vpcCIDR"),
		Env:                ctx.Stack(),
		VpcTagKey:          "Name",
		PublicSubnetCirdr:  conf.Require("publicSubnetCIDR"),
		PrivateSubnetCirdr: conf.Require("privateSubnetCIDR"),
	}

	err := parameters.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate parameters: %w", err)
	}

	Azs, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{
		State: pulumi.StringRef("available")})

	if err != nil {
		return fmt.Errorf("failed getting AZs: %w", err)
	}

	existingVpcs, err := ec2.GetVpcs(ctx, &ec2.GetVpcsArgs{
		Tags: map[string]string{
			parameters.VpcTagKey: parameters.Name,
		},
	})

	if err != nil {
		return fmt.Errorf("failed importing existing vpc: %w", err)
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
			return fmt.Errorf("failed importing existing VPC: %w", err)
		}

		ctx.Log.Info(fmt.Sprintf("VPC already exists with ID: %v", existingVpcs.Ids[0]), nil)

	} else {
		ctx.Log.Debug("VPC does not exist, creating a new one...", nil)

		vpc, err = ec2.NewVpc(ctx, parameters.Name, &ec2.VpcArgs{
			CidrBlock: pulumi.String(parameters.Cidr),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(parameters.Name),
			},
		})

		if err != nil {
			return fmt.Errorf("failed creating VPC: %w", err)
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
	}, pulumi.Parent(vpc))

	if err != nil {
		return fmt.Errorf("failed creating public subnet: %w", err)
	}

	privateSubnet, err := ec2.NewSubnet(ctx, "privateSubnet", &ec2.SubnetArgs{
		VpcId:            vpc.ID(),
		CidrBlock:        pulumi.String(parameters.PrivateSubnetCirdr),
		AvailabilityZone: pulumi.String(Azs.Names[1]),

		Tags: pulumi.StringMap{
			"Name": pulumi.Sprintf("%s-private-subnet-%s", parameters.Name, parameters.Env),
		},
	}, pulumi.Parent(vpc))

	if err != nil {
		return fmt.Errorf("failed creating private subnet: %w", err)
	}
	/**
		internetGateway, err := ec2.NewInternetGateway(ctx, "internetGateway", &ec2.InternetGatewayArgs{
			Tags: pulumi.StringMap{
				"Name": pulumi.Sprintf("%s-internet-gateway-%s", parameters.Name, parameters.Env),
			},
			VpcId: vpc.ID(),
		}, pulumi.Parent(vpc))
	**/
	ctx.Export("VpcId", vpc.ID())
	ctx.Export("PublicSubnetIds", publicSubnet.ID())
	ctx.Export("PrivateSubnetIds", privateSubnet.ID())
	ctx.Export("PublicSubnetAz", publicSubnet.AvailabilityZone)
	ctx.Export("PrivateSubnetAz", privateSubnet.AvailabilityZone)

	return nil
}
