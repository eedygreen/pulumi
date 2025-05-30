package vpc_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/eedygreen/pulumi/infra/vpc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

func TestValidateParams(t *testing.T) {

	testCases := []struct {
		name        string
		params      map[string]string
		expectedErr string
		//expected    vpc.Parameters
	}{
		{

			// valid params
			name: "valid params",
			params: map[string]string{
				"vpcName":           "test-vpc",
				"vpcCidr":           "10.0.0.0/16",
				"Env":               "dev",
				"VpcTagKey":         "Name",
				"publicSubnetCidr":  "10.0.1.0/24",
				"privateSubnetCidr": "10.0.2.0/24",
			},
			expectedErr: "",
		},
		{
			name: "vpcName params missing",
			params: map[string]string{
				"vpcCidr":           "10.0.0.0/16",
				"Env":               "dev",
				"VpcTagKey":         "Name",
				"publicSubnetCidr":  "10.0.1.0/24",
				"privateSubnetCidr": "10.0.2.0/24",
			},
			expectedErr: "VpcName is required in pulumi config",
		},
		{
			name: "vpcCidr params missing",
			params: map[string]string{
				"vpcName":           "test-vpc",
				"Env":               "dev",
				"VpcTagKey":         "Name",
				"publicSubnetCidr":  "10.0.1.0/24",
				"privateSubnetCidr": "10.0.2.0/24",
			},
			expectedErr: "VpcCidr is required in pulumi config",
		},
		{
			name: "publicsubnet params missing",
			params: map[string]string{
				"vpcName":           "test-vpc",
				"Env":               "dev",
				"VpcTagKey":         "Name",
				"vpcCidr":           "10.0.0.0/16",
				"privateSubnetCidr": "10.0.2.0/24",
			},
			expectedErr: "publicsubnet is required in pulumi config",
		},
		{
			name: "privatesubnet params missing",
			params: map[string]string{
				"vpcName":          "test-vpc",
				"Env":              "dev",
				"VpcTagKey":        "Name",
				"vpcCidr":          "10.0.0.0/16",
				"publicSubnetCidr": "10.0.1.0/24",
			},
			expectedErr: "privatesubnet is required in pulumi config",
		},
		{
			name:        "all params missing",
			params:      map[string]string{},
			expectedErr: "vpcName is required in pulumi config",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				//cfg := config.New(ctx, "")
				//cfg.Stack()
				for key, value := range test.params {
					os.Setenv(fmt.Sprintf("PULUMI_CONFIG_%s", key), value)
					defer os.Unsetenv(fmt.Sprintf("PULUMI_CONFIG_%s", key))
				}

				data := &vpc.Parameters{}
				_, err := data.ValidateParams(ctx)

				if test.expectedErr == "" {
					assert.Error(t, err)

				} else {
					assert.EqualError(t, err, test.expectedErr)
				}
				return nil
			}, pulumi.WithMocks("pulumi", "dev", nil))

			assert.NoError(t, err, "Pulumi Run Failed")
		})
	}
}
