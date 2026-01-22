package resources

import (
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("securityGroup", CollectSecurityGroups)
}

// SecurityGroup represents an Alibaba Cloud Security Group resource
type SecurityGroup struct {
	Client *ecs.Client
	Region string
}

// CollectSecurityGroups discovers all Security Groups in the specified region
func CollectSecurityGroups(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateECSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allSecurityGroups []*ecs.DescribeSecurityGroupsResponseBodySecurityGroupsSecurityGroup
	pageNumber := int32(1)
	pageSize := int32(50)

	// Paginate through all security groups
	for {
		request := &ecs.DescribeSecurityGroupsRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeSecurityGroups(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.SecurityGroups != nil && response.Body.SecurityGroups.SecurityGroup != nil {
			allSecurityGroups = append(allSecurityGroups, response.Body.SecurityGroups.SecurityGroup...)
		}

		// Check if we've retrieved all security groups
		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allSecurityGroups)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, sg := range allSecurityGroups {
		sgID := ""
		if sg.SecurityGroupId != nil {
			sgID = *sg.SecurityGroupId
		}

		sgName := ""
		if sg.SecurityGroupName != nil {
			sgName = *sg.SecurityGroupName
		}
		// If no name, use Security Group ID as the name for display
		if sgName == "" {
			sgName = sgID
		}

		res := types.Resource{
			Removable:    SecurityGroup{Client: client, Region: region},
			Region:       region,
			ResourceID:   sgID,
			ResourceName: sgName,
			ProductName:  "SecurityGroup",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Security Group
func (sg SecurityGroup) Remove(region string, resourceID string, resourceName string) error {
	request := &ecs.DeleteSecurityGroupRequest{
		SecurityGroupId: tea.String(resourceID),
		RegionId:        tea.String(region),
	}

	_, err := sg.Client.DeleteSecurityGroup(request)
	return err
}
