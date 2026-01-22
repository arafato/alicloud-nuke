package utils

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	nas "github.com/alibabacloud-go/nas-20170626/v3/client"
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

// CreateNASClient creates a NAS client for a specific region
func CreateNASClient(creds *types.Credentials, region string) (*nas.Client, error) {
	// NAS requires a region-specific endpoint
	endpoint := "nas." + region + ".aliyuncs.com"
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String(region),
		Endpoint:        tea.String(endpoint),
	}

	return nas.NewClient(config)
}
