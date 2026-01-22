package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("routerInterface", CollectRouterInterfaces)
}

// RouterInterface represents an Alibaba Cloud Router Interface resource (used for VPC peering)
type RouterInterface struct {
	Client *vpc.Client
	Region string
}

// CollectRouterInterfaces discovers all Router Interfaces in the specified region
func CollectRouterInterfaces(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allRouterInterfaces []*vpc.DescribeRouterInterfacesResponseBodyRouterInterfaceSetRouterInterfaceType
	pageNumber := int32(1)
	pageSize := int32(50)

	// Paginate through all router interfaces
	for {
		request := &vpc.DescribeRouterInterfacesRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeRouterInterfaces(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.RouterInterfaceSet != nil && response.Body.RouterInterfaceSet.RouterInterfaceType != nil {
			allRouterInterfaces = append(allRouterInterfaces, response.Body.RouterInterfaceSet.RouterInterfaceType...)
		}

		// Check if we've retrieved all router interfaces
		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allRouterInterfaces)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, ri := range allRouterInterfaces {
		riID := ""
		if ri.RouterInterfaceId != nil {
			riID = *ri.RouterInterfaceId
		}

		riName := ""
		if ri.Name != nil {
			riName = *ri.Name
		}
		// If no name, use Router Interface ID as the name for display
		if riName == "" {
			riName = riID
		}

		// Add role info to help identify the interface
		role := ""
		if ri.Role != nil {
			role = *ri.Role
		}
		if role != "" && riName == riID {
			riName = riID + " (" + role + ")"
		}

		res := types.Resource{
			Removable:    RouterInterface{Client: client, Region: region},
			Region:       region,
			ResourceID:   riID,
			ResourceName: riName,
			ProductName:  "RouterInterface",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Router Interface
func (ri RouterInterface) Remove(region string, resourceID string, resourceName string) error {
	// First deactivate the router interface if it's active
	deactivateReq := &vpc.DeactivateRouterInterfaceRequest{
		RouterInterfaceId: tea.String(resourceID),
		RegionId:          tea.String(region),
	}
	// Ignore deactivation errors - it may already be inactive
	ri.Client.DeactivateRouterInterface(deactivateReq)

	// Delete the router interface
	request := &vpc.DeleteRouterInterfaceRequest{
		RouterInterfaceId: tea.String(resourceID),
		RegionId:          tea.String(region),
	}

	_, err := ri.Client.DeleteRouterInterface(request)
	return err
}
