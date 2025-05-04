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

func Validate(ctx *pulumi.Context, opts ...Parameters) (*Parameters, error) {
	conf := config.New(ctx, "")
	p := Parameters{
		Name:               conf.Require("vpcName"),
		Cidr:               conf.Require("vpcCIDR"),
		Env:                ctx.Stack(),
		VpcTagKey:          "Name",
		PublicSubnetCirdr:  conf.Require("publicSubnetCIDR"),
		PrivateSubnetCirdr: conf.Require("privateSubnetCIDR"),
	}

	switch {
	case p.Name == "":
		return nil, fmt.Errorf("vpcName is required in pulumi config")

	case p.Cidr == "":
		return nil, fmt.Errorf("vpcCIDR is required in pulumi config")

	case p.PublicSubnetCirdr == "":
		return nil, fmt.Errorf("publicSubnetCIDR is required in pulumi config")

	case p.PrivateSubnetCirdr == "":
		return nil, fmt.Errorf("privateSubnetCIDR is required in pulumi config")

	}
	return &p, nil
}

func CreateSubnets(ctx *pulumi.Context, vpc *ec2.Vpc, p *Parameters) ([]*ec2.Subnet, error) {

	Azs, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{
		State: pulumi.StringRef("available")})

	if err != nil {
		return nil, fmt.Errorf("failed getting AZs: %w", err)
	}

	publicSubnet, err := ec2.NewSubnet(ctx, "publicSubnet", &ec2.SubnetArgs{
		VpcId:            vpc.ID(),
		CidrBlock:        pulumi.String(p.PublicSubnetCirdr),
		AvailabilityZone: pulumi.String(Azs.Names[0]),
		Tags: pulumi.StringMap{
			"Name": pulumi.Sprintf("%s-public-subnet-%s", p.Name, p.Env),
		},
	}, pulumi.Parent(vpc))

	if err != nil {
		return nil, fmt.Errorf("failed creating public subnet: %w", err)
	}

	privateSubnet, err := ec2.NewSubnet(ctx, "privateSubnet", &ec2.SubnetArgs{
		VpcId:            vpc.ID(),
		CidrBlock:        pulumi.String(p.PrivateSubnetCirdr),
		AvailabilityZone: pulumi.String(Azs.Names[1]),

		Tags: pulumi.StringMap{
			"Name": pulumi.Sprintf("%s-private-subnet-%s", p.Name, p.Env),
		},
	}, pulumi.Parent(vpc))

	if err != nil {
		return nil, fmt.Errorf("failed creating private subnet: %w", err)
	}

	return []*ec2.Subnet{publicSubnet, privateSubnet}, nil
}

func CreateVpc(ctx *pulumi.Context, parameters *Parameters) (*ec2.Vpc, error) {

	existingVpcs, err := ec2.GetVpcs(ctx, &ec2.GetVpcsArgs{
		Tags: map[string]string{
			parameters.VpcTagKey: parameters.Name,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed importing existing vpc: %w", err)
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
			return vpc, fmt.Errorf("failed importing existing VPC: %w", err)
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
			return vpc, fmt.Errorf("failed creating VPC: %w", err)
		}
		ctx.Log.Info(fmt.Sprintf("Created VPC with ID: %v", vpc.ID()), nil)

	}

	/**
		internetGateway, err := ec2.NewInternetGateway(ctx, "internetGateway", &ec2.InternetGatewayArgs{
			Tags: pulumi.StringMap{
				"Name": pulumi.Sprintf("%s-internet-gateway-%s", parameters.Name, parameters.Env),
			},
			VpcId: vpc.ID(),
		}, pulumi.Parent(vpc))
	**/

	return vpc, nil
}
