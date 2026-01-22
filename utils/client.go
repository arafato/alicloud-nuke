package utils

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/types"
)

// CreateECSClient creates an ECS client for a specific region
func CreateECSClient(creds *types.Credentials, region string) (*ecs.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String(region),
	}

	return ecs.NewClient(config)
}

// CreateVPCClient creates a VPC client for a specific region
func CreateVPCClient(creds *types.Credentials, region string) (*vpc.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String(region),
	}

	return vpc.NewClient(config)
}
