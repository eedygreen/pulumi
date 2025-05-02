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
		Cidr:               config.New(ctx, "").Get("vpcCIDR"),
		Env:                ctx.Stack(),
		VpcTagKey:          "Name",
		PublicSubnetCirdr:  config.New(ctx, "").Get("publicSubnetCIDR"),
		PrivateSubnetCirdr: config.New(ctx, "").Get("privateSubnetCIDR"),
	}

	Azs, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{
		State: pulumi.StringRef("available")}, nil)
	if err != nil {
		return pulumi.IDOutput{}, err
	}

	isVpcAlreadyExist, err := ec2.GetVpcs(ctx, &ec2.GetVpcsArgs{
		Tags: map[string]string{
			parameters.VpcTagKey: parameters.Name,
		},
	})

	if err != nil {
		return pulumi.IDOutput{}, err
	}
	var vpcId pulumi.IDOutput

	if len(isVpcAlreadyExist.Ids) > 0 {
		vpcId = pulumi.ID(isVpcAlreadyExist.Ids[0]).ToIDOutput()
		fmt.Println("VPC already exists with ID: ", pulumi.StringOutput(vpcId))
		return vpcId, nil
	} else {

		fmt.Println("VPC does not exist, creating a new one...")

		vpc, err := ec2.NewVpc(ctx, parameters.Name, &ec2.VpcArgs{
			CidrBlock: pulumi.StringPtr(parameters.Cidr),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(parameters.Name),
			},
		})

		vpcId = vpc.ID()

		if err != nil {
			return pulumi.IDOutput{}, err
		}
		fmt.Println("Created VPC with ID: ", vpcId)
	}

	publicSubnet, err := ec2.NewSubnet(ctx, "publicSubnet", &ec2.SubnetArgs{
		VpcId:            vpcId,
		CidrBlock:        pulumi.StringPtr(parameters.PublicSubnetCirdr),
		AvailabilityZone: pulumi.String(Azs.Names[0]),
		Tags: pulumi.StringMap{
			"Name": pulumi.Sprintf("%s-public-subnet-%s", parameters.Name, parameters.Env),
		},
	})

	if err != nil {
		return pulumi.IDOutput{}, err
	}

	privateSubnet, err := ec2.NewSubnet(ctx, "privateSubnet", &ec2.SubnetArgs{
		VpcId:            vpcId,
		CidrBlock:        pulumi.StringPtr(parameters.PrivateSubnetCirdr),
		AvailabilityZone: pulumi.String(Azs.Names[1]),

		Tags: pulumi.StringMap{
			"Name": pulumi.Sprintf("%s-public-subnet-%s", parameters.Name, parameters.Env),
		},
	})

	if err != nil {
		return pulumi.IDOutput{}, err
	}

	ctx.Export("VpcId", vpcId)
	ctx.Export("PublicSubnetIds", publicSubnet.ID())
	ctx.Export("PrivateSubnetIds", privateSubnet.ID())
	ctx.Export("PublicSubnetAz", publicSubnet.AvailabilityZone)
	ctx.Export("PrivateSubnetAz", privateSubnet.AvailabilityZone)

	return vpcId, nil
}
