package utils

import (
	alb "github.com/alibabacloud-go/alb-20200616/v2/client"
	cbn "github.com/alibabacloud-go/cbn-20170912/v2/client"
	cr "github.com/alibabacloud-go/cr-20181201/v2/client"
	cs "github.com/alibabacloud-go/cs-20151215/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dds "github.com/alibabacloud-go/dds-20151201/v4/client"
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	ess "github.com/alibabacloud-go/ess-20220222/v2/client"
	nas "github.com/alibabacloud-go/nas-20170626/v3/client"
	nlb "github.com/alibabacloud-go/nlb-20220430/v2/client"
	polardb "github.com/alibabacloud-go/polardb-20170801/v6/client"
	r_kvstore "github.com/alibabacloud-go/r-kvstore-20150101/v4/client"
	rds "github.com/alibabacloud-go/rds-20140815/v4/client"
	slb "github.com/alibabacloud-go/slb-20140515/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

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

// CreateESSClient creates an Auto Scaling (ESS) client for a specific region
func CreateESSClient(creds *types.Credentials, region string) (*ess.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String(region),
	}
	return ess.NewClient(config)
}

// CreateCRClient creates a Container Registry client for a specific region
func CreateCRClient(creds *types.Credentials, region string) (*cr.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String(region),
	}
	return cr.NewClient(config)
}

// CreateSLBClient creates a Classic Load Balancer (SLB) client for a specific region
func CreateSLBClient(creds *types.Credentials, region string) (*slb.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String(region),
	}
	return slb.NewClient(config)
}

// CreateALBClient creates an Application Load Balancer (ALB) client for a specific region
func CreateALBClient(creds *types.Credentials, region string) (*alb.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String(region),
	}
	return alb.NewClient(config)
}

// CreateNLBClient creates a Network Load Balancer (NLB) client for a specific region
func CreateNLBClient(creds *types.Credentials, region string) (*nlb.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String(region),
	}
	return nlb.NewClient(config)
}

// CreateRDSClient creates an RDS client for a specific region
func CreateRDSClient(creds *types.Credentials, region string) (*rds.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String(region),
	}
	return rds.NewClient(config)
}

// CreateRedisClient creates a Redis (KVStore) client for a specific region
func CreateRedisClient(creds *types.Credentials, region string) (*r_kvstore.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String(region),
	}
	return r_kvstore.NewClient(config)
}

// CreateMongoDBClient creates a MongoDB (DDS) client for a specific region
func CreateMongoDBClient(creds *types.Credentials, region string) (*dds.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String(region),
	}
	return dds.NewClient(config)
}

// CreatePolarDBClient creates a PolarDB client for a specific region
func CreatePolarDBClient(creds *types.Credentials, region string) (*polardb.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String(region),
	}
	return polardb.NewClient(config)
}

// CreateOSSClient creates an OSS client for a specific region
func CreateOSSClient(creds *types.Credentials, region string) (*oss.Client, error) {
	provider := credentials.NewStaticCredentialsProvider(creds.AccessKeyID, creds.AccessKeySecret)
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(provider).
		WithRegion(region)

	return oss.NewClient(cfg), nil
}

// CreateCSClient creates a Container Service (ACK) client for a specific region
func CreateCSClient(creds *types.Credentials, region string) (*cs.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String(region),
	}
	return cs.NewClient(config)
}

// CreateCENClient creates a Cloud Enterprise Network (CEN) client for a specific region
func CreateCENClient(creds *types.Credentials, region string) (*cbn.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String(region),
	}
	return cbn.NewClient(config)
}
