package resources

import (
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("keyPair", CollectKeyPairs)
}

// KeyPair represents an Alibaba Cloud ECS Key Pair resource
type KeyPair struct {
	Client *ecs.Client
	Region string
}

// CollectKeyPairs discovers all Key Pairs in the specified region
func CollectKeyPairs(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateECSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allKeyPairs []*ecs.DescribeKeyPairsResponseBodyKeyPairsKeyPair
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &ecs.DescribeKeyPairsRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeKeyPairs(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.KeyPairs != nil && response.Body.KeyPairs.KeyPair != nil {
			allKeyPairs = append(allKeyPairs, response.Body.KeyPairs.KeyPair...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allKeyPairs)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, keyPair := range allKeyPairs {
		keyPairName := ""
		if keyPair.KeyPairName != nil {
			keyPairName = *keyPair.KeyPairName
		}

		// Key pairs use name as the identifier
		res := types.Resource{
			Removable:    KeyPair{Client: client, Region: region},
			Region:       region,
			ResourceID:   keyPairName,
			ResourceName: keyPairName,
			ProductName:  "KeyPair",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Key Pair
func (k KeyPair) Remove(region string, resourceID string, resourceName string) error {
	request := &ecs.DeleteKeyPairsRequest{
		RegionId:     tea.String(region),
		KeyPairNames: tea.String("[\"" + resourceID + "\"]"), // API expects JSON array
	}

	_, err := k.Client.DeleteKeyPairs(request)
	return err
}
