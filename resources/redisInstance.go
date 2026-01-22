package resources

import (
	r_kvstore "github.com/alibabacloud-go/r-kvstore-20150101/v4/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("redisInstance", CollectRedisInstances)
}

// RedisInstance represents an Alibaba Cloud Redis Instance resource
type RedisInstance struct {
	Client *r_kvstore.Client
	Region string
}

// CollectRedisInstances discovers all Redis instances in the specified region
func CollectRedisInstances(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateRedisClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allInstances []*r_kvstore.DescribeInstancesResponseBodyInstancesKVStoreInstance
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &r_kvstore.DescribeInstancesRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeInstances(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.Instances != nil && response.Body.Instances.KVStoreInstance != nil {
			allInstances = append(allInstances, response.Body.Instances.KVStoreInstance...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allInstances)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, instance := range allInstances {
		instanceID := ""
		if instance.InstanceId != nil {
			instanceID = *instance.InstanceId
		}

		instanceName := ""
		if instance.InstanceName != nil {
			instanceName = *instance.InstanceName
		}
		if instanceName == "" {
			instanceName = instanceID
		}

		res := types.Resource{
			Removable:    RedisInstance{Client: client, Region: region},
			Region:       region,
			ResourceID:   instanceID,
			ResourceName: instanceName,
			ProductName:  "RedisInstance",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Redis instance
func (r RedisInstance) Remove(region string, resourceID string, resourceName string) error {
	request := &r_kvstore.DeleteInstanceRequest{
		InstanceId: tea.String(resourceID),
	}

	_, err := r.Client.DeleteInstance(request)
	return err
}
