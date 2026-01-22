package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("sslVpnServer", CollectSslVpnServers)
}

// SslVpnServer represents an Alibaba Cloud SSL VPN Server resource
type SslVpnServer struct {
	Client *vpc.Client
	Region string
}

// CollectSslVpnServers discovers all SSL VPN Servers in the specified region
func CollectSslVpnServers(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allSslVpnServers []*vpc.DescribeSslVpnServersResponseBodySslVpnServersSslVpnServer
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &vpc.DescribeSslVpnServersRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeSslVpnServers(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.SslVpnServers != nil && response.Body.SslVpnServers.SslVpnServer != nil {
			allSslVpnServers = append(allSslVpnServers, response.Body.SslVpnServers.SslVpnServer...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allSslVpnServers)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, server := range allSslVpnServers {
		serverID := ""
		if server.SslVpnServerId != nil {
			serverID = *server.SslVpnServerId
		}

		serverName := ""
		if server.Name != nil {
			serverName = *server.Name
		}
		if serverName == "" {
			serverName = serverID
		}

		res := types.Resource{
			Removable:    SslVpnServer{Client: client, Region: region},
			Region:       region,
			ResourceID:   serverID,
			ResourceName: serverName,
			ProductName:  "SslVpnServer",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the SSL VPN Server
func (s SslVpnServer) Remove(region string, resourceID string, resourceName string) error {
	request := &vpc.DeleteSslVpnServerRequest{
		SslVpnServerId: tea.String(resourceID),
		RegionId:       tea.String(region),
	}

	_, err := s.Client.DeleteSslVpnServer(request)
	return err
}
