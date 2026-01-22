package resources

import (
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("networkInterface", CollectNetworkInterfaces)
}

// NetworkInterface represents an Alibaba Cloud Elastic Network Interface (ENI) resource
type NetworkInterface struct {
	Client *ecs.Client
	Region string
}

// CollectNetworkInterfaces discovers all Network Interfaces in the specified region
func CollectNetworkInterfaces(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateECSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allENIs []*ecs.DescribeNetworkInterfacesResponseBodyNetworkInterfaceSetsNetworkInterfaceSet
	pageNumber := int32(1)
	pageSize := int32(100)

	// Paginate through all network interfaces
	for {
		request := &ecs.DescribeNetworkInterfacesRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeNetworkInterfaces(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.NetworkInterfaceSets != nil && response.Body.NetworkInterfaceSets.NetworkInterfaceSet != nil {
			allENIs = append(allENIs, response.Body.NetworkInterfaceSets.NetworkInterfaceSet...)
		}

		// Check if we've retrieved all network interfaces
		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allENIs)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, eni := range allENIs {
		eniID := ""
		if eni.NetworkInterfaceId != nil {
			eniID = *eni.NetworkInterfaceId
		}

		eniName := ""
		if eni.NetworkInterfaceName != nil {
			eniName = *eni.NetworkInterfaceName
		}
		// If no name, use ENI ID as the name for display
		if eniName == "" {
			eniName = eniID
		}

		// Skip primary ENIs - they are deleted with the ECS instance
		eniType := ""
		if eni.Type != nil {
			eniType = *eni.Type
		}
		if eniType == "Primary" {
			continue
		}

		res := types.Resource{
			Removable:    NetworkInterface{Client: client, Region: region},
			Region:       region,
			ResourceID:   eniID,
			ResourceName: eniName,
			ProductName:  "NetworkInterface",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Network Interface
func (eni NetworkInterface) Remove(region string, resourceID string, resourceName string) error {
	// First detach the ENI if it's attached to an instance
	detachReq := &ecs.DetachNetworkInterfaceRequest{
		RegionId:           tea.String(region),
		NetworkInterfaceId: tea.String(resourceID),
	}
	// Ignore detach errors - it may not be attached
	eni.Client.DetachNetworkInterface(detachReq)

	// Delete the network interface
	request := &ecs.DeleteNetworkInterfaceRequest{
		RegionId:           tea.String(region),
		NetworkInterfaceId: tea.String(resourceID),
	}

	_, err := eni.Client.DeleteNetworkInterface(request)
	return err
}
