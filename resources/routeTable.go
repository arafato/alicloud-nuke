package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("routeTable", CollectRouteTables)
}

// RouteTable represents an Alibaba Cloud Route Table resource
type RouteTable struct {
	Client *vpc.Client
	Region string
}

// CollectRouteTables discovers all Route Tables in the specified region
func CollectRouteTables(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allRouteTables []*vpc.DescribeRouteTableListResponseBodyRouterTableListRouterTableListType
	pageNumber := int32(1)
	pageSize := int32(50)

	// Paginate through all route tables
	for {
		request := &vpc.DescribeRouteTableListRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeRouteTableList(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.RouterTableList != nil && response.Body.RouterTableList.RouterTableListType != nil {
			allRouteTables = append(allRouteTables, response.Body.RouterTableList.RouterTableListType...)
		}

		// Check if we've retrieved all route tables
		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allRouteTables)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, rt := range allRouteTables {
		rtID := ""
		if rt.RouteTableId != nil {
			rtID = *rt.RouteTableId
		}

		rtName := ""
		if rt.RouteTableName != nil {
			rtName = *rt.RouteTableName
		}
		// If no name, use Route Table ID as the name for display
		if rtName == "" {
			rtName = rtID
		}

		// Check if this is a system route table (cannot be deleted)
		// System route tables have RouteTableType = "System"
		isSystem := false
		if rt.RouteTableType != nil && *rt.RouteTableType == "System" {
			isSystem = true
		}

		res := types.Resource{
			Removable:    RouteTable{Client: client, Region: region},
			Region:       region,
			ResourceID:   rtID,
			ResourceName: rtName,
			ProductName:  "RouteTable",
		}

		// Hide system route tables - they cannot be deleted
		if isSystem {
			res.SetState(types.Hidden)
		}

		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Route Table
func (rt RouteTable) Remove(region string, resourceID string, resourceName string) error {
	request := &vpc.DeleteRouteTableRequest{
		RouteTableId: tea.String(resourceID),
		RegionId:     tea.String(region),
	}

	_, err := rt.Client.DeleteRouteTable(request)
	return err
}
