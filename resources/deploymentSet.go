package resources

import (
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("deploymentSet", CollectDeploymentSets)
}

// DeploymentSet represents an Alibaba Cloud ECS Deployment Set resource
type DeploymentSet struct {
	Client *ecs.Client
	Region string
}

// CollectDeploymentSets discovers all Deployment Sets in the specified region
func CollectDeploymentSets(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateECSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allDeploymentSets []*ecs.DescribeDeploymentSetsResponseBodyDeploymentSetsDeploymentSet
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &ecs.DescribeDeploymentSetsRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeDeploymentSets(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.DeploymentSets != nil && response.Body.DeploymentSets.DeploymentSet != nil {
			allDeploymentSets = append(allDeploymentSets, response.Body.DeploymentSets.DeploymentSet...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allDeploymentSets)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, ds := range allDeploymentSets {
		dsID := ""
		if ds.DeploymentSetId != nil {
			dsID = *ds.DeploymentSetId
		}

		dsName := ""
		if ds.DeploymentSetName != nil {
			dsName = *ds.DeploymentSetName
		}
		if dsName == "" {
			dsName = dsID
		}

		res := types.Resource{
			Removable:    DeploymentSet{Client: client, Region: region},
			Region:       region,
			ResourceID:   dsID,
			ResourceName: dsName,
			ProductName:  "DeploymentSet",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Deployment Set
func (d DeploymentSet) Remove(region string, resourceID string, resourceName string) error {
	request := &ecs.DeleteDeploymentSetRequest{
		RegionId:        tea.String(region),
		DeploymentSetId: tea.String(resourceID),
	}

	_, err := d.Client.DeleteDeploymentSet(request)
	return err
}
