package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("vpnConnection", CollectVpnConnections)
}

// VpnConnection represents an Alibaba Cloud VPN Connection (IPsec Connection) resource
type VpnConnection struct {
	Client *vpc.Client
	Region string
}

// CollectVpnConnections discovers all VPN Connections in the specified region
func CollectVpnConnections(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allVpnConnections []*vpc.DescribeVpnConnectionsResponseBodyVpnConnectionsVpnConnection
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &vpc.DescribeVpnConnectionsRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeVpnConnections(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.VpnConnections != nil && response.Body.VpnConnections.VpnConnection != nil {
			allVpnConnections = append(allVpnConnections, response.Body.VpnConnections.VpnConnection...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allVpnConnections)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, conn := range allVpnConnections {
		connID := ""
		if conn.VpnConnectionId != nil {
			connID = *conn.VpnConnectionId
		}

		connName := ""
		if conn.Name != nil {
			connName = *conn.Name
		}
		if connName == "" {
			connName = connID
		}

		res := types.Resource{
			Removable:    VpnConnection{Client: client, Region: region},
			Region:       region,
			ResourceID:   connID,
			ResourceName: connName,
			ProductName:  "VpnConnection",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the VPN Connection
func (v VpnConnection) Remove(region string, resourceID string, resourceName string) error {
	request := &vpc.DeleteVpnConnectionRequest{
		VpnConnectionId: tea.String(resourceID),
		RegionId:        tea.String(region),
	}

	_, err := v.Client.DeleteVpnConnection(request)
	return err
}
