package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("natGateway", CollectNatGateways)
}

// NatGateway represents an Alibaba Cloud NAT Gateway resource
type NatGateway struct {
	Client *vpc.Client
	Region string
}

// CollectNatGateways discovers all NAT Gateways in the specified region
func CollectNatGateways(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allNatGateways []*vpc.DescribeNatGatewaysResponseBodyNatGatewaysNatGateway
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &vpc.DescribeNatGatewaysRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeNatGateways(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.NatGateways != nil && response.Body.NatGateways.NatGateway != nil {
			allNatGateways = append(allNatGateways, response.Body.NatGateways.NatGateway...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allNatGateways)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, nat := range allNatGateways {
		natID := ""
		if nat.NatGatewayId != nil {
			natID = *nat.NatGatewayId
		}

		natName := ""
		if nat.Name != nil {
			natName = *nat.Name
		}
		if natName == "" {
			natName = natID
		}

		res := types.Resource{
			Removable:    NatGateway{Client: client, Region: region},
			Region:       region,
			ResourceID:   natID,
			ResourceName: natName,
			ProductName:  "NatGateway",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the NAT Gateway
func (n NatGateway) Remove(region string, resourceID string, resourceName string) error {
	request := &vpc.DeleteNatGatewayRequest{
		NatGatewayId: tea.String(resourceID),
		RegionId:     tea.String(region),
		Force:        tea.Bool(true), // Force delete even if SNAT/DNAT entries exist
	}

	_, err := n.Client.DeleteNatGateway(request)
	return err
}
