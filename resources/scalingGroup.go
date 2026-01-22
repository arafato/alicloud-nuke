package resources

import (
	ess "github.com/alibabacloud-go/ess-20220222/v2/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("scalingGroup", CollectScalingGroups)
}

// ScalingGroup represents an Alibaba Cloud Auto Scaling Group resource
type ScalingGroup struct {
	Client *ess.Client
	Region string
}

// CollectScalingGroups discovers all Auto Scaling Groups in the specified region
func CollectScalingGroups(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateESSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allGroups []*ess.DescribeScalingGroupsResponseBodyScalingGroups
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &ess.DescribeScalingGroupsRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeScalingGroups(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.ScalingGroups != nil {
			allGroups = append(allGroups, response.Body.ScalingGroups...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allGroups)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, group := range allGroups {
		groupID := ""
		if group.ScalingGroupId != nil {
			groupID = *group.ScalingGroupId
		}

		groupName := ""
		if group.ScalingGroupName != nil {
			groupName = *group.ScalingGroupName
		}
		if groupName == "" {
			groupName = groupID
		}

		res := types.Resource{
			Removable:    ScalingGroup{Client: client, Region: region},
			Region:       region,
			ResourceID:   groupID,
			ResourceName: groupName,
			ProductName:  "ScalingGroup",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Scaling Group
func (s ScalingGroup) Remove(region string, resourceID string, resourceName string) error {
	// First disable the scaling group
	disableReq := &ess.DisableScalingGroupRequest{
		ScalingGroupId: tea.String(resourceID),
	}
	s.Client.DisableScalingGroup(disableReq)

	// Delete the scaling group with force
	request := &ess.DeleteScalingGroupRequest{
		ScalingGroupId: tea.String(resourceID),
		ForceDelete:    tea.Bool(true), // Force delete instances in the group
	}

	_, err := s.Client.DeleteScalingGroup(request)
	return err
}
