package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("customerGateway", CollectCustomerGateways)
}

// CustomerGateway represents an Alibaba Cloud Customer Gateway resource (for VPN)
type CustomerGateway struct {
	Client *vpc.Client
	Region string
}

// CollectCustomerGateways discovers all Customer Gateways in the specified region
func CollectCustomerGateways(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allCustomerGateways []*vpc.DescribeCustomerGatewaysResponseBodyCustomerGatewaysCustomerGateway
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &vpc.DescribeCustomerGatewaysRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeCustomerGateways(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.CustomerGateways != nil && response.Body.CustomerGateways.CustomerGateway != nil {
			allCustomerGateways = append(allCustomerGateways, response.Body.CustomerGateways.CustomerGateway...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allCustomerGateways)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, cgw := range allCustomerGateways {
		cgwID := ""
		if cgw.CustomerGatewayId != nil {
			cgwID = *cgw.CustomerGatewayId
		}

		cgwName := ""
		if cgw.Name != nil {
			cgwName = *cgw.Name
		}
		if cgwName == "" && cgw.IpAddress != nil {
			cgwName = *cgw.IpAddress // Use IP if no name
		}
		if cgwName == "" {
			cgwName = cgwID
		}

		res := types.Resource{
			Removable:    CustomerGateway{Client: client, Region: region},
			Region:       region,
			ResourceID:   cgwID,
			ResourceName: cgwName,
			ProductName:  "CustomerGateway",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Customer Gateway
func (c CustomerGateway) Remove(region string, resourceID string, resourceName string) error {
	request := &vpc.DeleteCustomerGatewayRequest{
		CustomerGatewayId: tea.String(resourceID),
		RegionId:          tea.String(region),
	}

	_, err := c.Client.DeleteCustomerGateway(request)
	return err
}
