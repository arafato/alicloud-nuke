package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("vpnGateway", CollectVpnGateways)
}

// VpnGateway represents an Alibaba Cloud VPN Gateway resource
type VpnGateway struct {
	Client *vpc.Client
	Region string
}

// CollectVpnGateways discovers all VPN Gateways in the specified region
func CollectVpnGateways(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allVpnGateways []*vpc.DescribeVpnGatewaysResponseBodyVpnGatewaysVpnGateway
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &vpc.DescribeVpnGatewaysRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeVpnGateways(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.VpnGateways != nil && response.Body.VpnGateways.VpnGateway != nil {
			allVpnGateways = append(allVpnGateways, response.Body.VpnGateways.VpnGateway...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allVpnGateways)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, vpn := range allVpnGateways {
		vpnID := ""
		if vpn.VpnGatewayId != nil {
			vpnID = *vpn.VpnGatewayId
		}

		vpnName := ""
		if vpn.Name != nil {
			vpnName = *vpn.Name
		}
		if vpnName == "" {
			vpnName = vpnID
		}

		res := types.Resource{
			Removable:    VpnGateway{Client: client, Region: region},
			Region:       region,
			ResourceID:   vpnID,
			ResourceName: vpnName,
			ProductName:  "VpnGateway",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the VPN Gateway
func (v VpnGateway) Remove(region string, resourceID string, resourceName string) error {
	request := &vpc.DeleteVpnGatewayRequest{
		VpnGatewayId: tea.String(resourceID),
		RegionId:     tea.String(region),
	}

	_, err := v.Client.DeleteVpnGateway(request)
	return err
}
