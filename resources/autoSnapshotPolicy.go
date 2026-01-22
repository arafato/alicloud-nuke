package resources

import (
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("autoSnapshotPolicy", CollectAutoSnapshotPolicies)
}

// AutoSnapshotPolicy represents an Alibaba Cloud ECS Auto Snapshot Policy resource
type AutoSnapshotPolicy struct {
	Client *ecs.Client
	Region string
}

// CollectAutoSnapshotPolicies discovers all Auto Snapshot Policies in the specified region
func CollectAutoSnapshotPolicies(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateECSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allPolicies []*ecs.DescribeAutoSnapshotPolicyExResponseBodyAutoSnapshotPoliciesAutoSnapshotPolicy
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &ecs.DescribeAutoSnapshotPolicyExRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeAutoSnapshotPolicyEx(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.AutoSnapshotPolicies != nil && response.Body.AutoSnapshotPolicies.AutoSnapshotPolicy != nil {
			allPolicies = append(allPolicies, response.Body.AutoSnapshotPolicies.AutoSnapshotPolicy...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allPolicies)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, policy := range allPolicies {
		policyID := ""
		if policy.AutoSnapshotPolicyId != nil {
			policyID = *policy.AutoSnapshotPolicyId
		}

		policyName := ""
		if policy.AutoSnapshotPolicyName != nil {
			policyName = *policy.AutoSnapshotPolicyName
		}
		if policyName == "" {
			policyName = policyID
		}

		res := types.Resource{
			Removable:    AutoSnapshotPolicy{Client: client, Region: region},
			Region:       region,
			ResourceID:   policyID,
			ResourceName: policyName,
			ProductName:  "AutoSnapshotPolicy",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Auto Snapshot Policy
func (a AutoSnapshotPolicy) Remove(region string, resourceID string, resourceName string) error {
	request := &ecs.DeleteAutoSnapshotPolicyRequest{
		RegionId:             tea.String(region),
		AutoSnapshotPolicyId: tea.String(resourceID),
	}

	_, err := a.Client.DeleteAutoSnapshotPolicy(request)
	return err
}
