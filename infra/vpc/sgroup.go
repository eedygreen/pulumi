package vpc

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type infra struct {
	group *ec2.SecurityGroup
}

func CreateSecurityGroup(ctx *pulumi.Context, name string) (*infra, error) {

	vpcSecurityGroup, err := ec2.NewSecurityGroup(ctx, name, &ec2.SecurityGroupArgs{
		Name: pulumi.String(name),
		Ingress: ec2.SecurityGroupIngressArray{
			ec2.SecurityGroupIngressArgs{
				Protocol: pulumi.String("tcp"),
				FromPort: pulumi.Int(0),
				ToPort:   pulumi.Int(65535),
				CidrBlocks: pulumi.StringArray{
					pulumi.String(""),
				},
			},
		},

		Egress: ec2.SecurityGroupEgressArray{},
	}, pulumi.Protect(false))

	if err != nil {
		return nil, fmt.Errorf("failed creating security group %s: %w", name, err)
	}

	return &infra{
		group: vpcSecurityGroup,
	}, nil
}
