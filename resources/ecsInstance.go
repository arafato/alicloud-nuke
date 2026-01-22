package resources

import (
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("ecsInstance", CollectECSInstances)
}

// ECSInstance represents an Alibaba Cloud ECS instance resource
type ECSInstance struct {
	Client *ecs.Client
	Region string
}

// CollectECSInstances discovers all ECS instances in the specified region
func CollectECSInstances(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateECSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allInstances []*ecs.DescribeInstancesResponseBodyInstancesInstance
	pageNumber := int32(1)
	pageSize := int32(100)

	// Paginate through all instances using PageNumber/PageSize
	for {
		request := &ecs.DescribeInstancesRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeInstances(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.Instances != nil && response.Body.Instances.Instance != nil {
			allInstances = append(allInstances, response.Body.Instances.Instance...)
		}

		// Check if we've retrieved all instances
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

		res := types.Resource{
			Removable:    ECSInstance{Client: client, Region: region},
			Region:       region,
			ResourceID:   instanceID,
			ResourceName: instanceName,
			ProductName:  "ECSInstance",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the ECS instance
func (e ECSInstance) Remove(region string, resourceID string, resourceName string) error {
	// Delete the instance with force option
	// Force=true allows deletion of running instances (will stop first) and subscription instances
	request := &ecs.DeleteInstanceRequest{
		InstanceId:            tea.String(resourceID),
		Force:                 tea.Bool(true),
		TerminateSubscription: tea.Bool(true),
	}

	_, err := e.Client.DeleteInstance(request)
	return err
}
