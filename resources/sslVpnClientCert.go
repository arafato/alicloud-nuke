package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("sslVpnClientCert", CollectSslVpnClientCerts)
}

// SslVpnClientCert represents an Alibaba Cloud SSL VPN Client Certificate resource
type SslVpnClientCert struct {
	Client *vpc.Client
	Region string
}

// CollectSslVpnClientCerts discovers all SSL VPN Client Certificates in the specified region
func CollectSslVpnClientCerts(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allCerts []*vpc.DescribeSslVpnClientCertsResponseBodySslVpnClientCertKeysSslVpnClientCertKey
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &vpc.DescribeSslVpnClientCertsRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeSslVpnClientCerts(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.SslVpnClientCertKeys != nil && response.Body.SslVpnClientCertKeys.SslVpnClientCertKey != nil {
			allCerts = append(allCerts, response.Body.SslVpnClientCertKeys.SslVpnClientCertKey...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allCerts)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, cert := range allCerts {
		certID := ""
		if cert.SslVpnClientCertId != nil {
			certID = *cert.SslVpnClientCertId
		}

		certName := ""
		if cert.Name != nil {
			certName = *cert.Name
		}
		if certName == "" {
			certName = certID
		}

		res := types.Resource{
			Removable:    SslVpnClientCert{Client: client, Region: region},
			Region:       region,
			ResourceID:   certID,
			ResourceName: certName,
			ProductName:  "SslVpnClientCert",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the SSL VPN Client Certificate
func (s SslVpnClientCert) Remove(region string, resourceID string, resourceName string) error {
	request := &vpc.DeleteSslVpnClientCertRequest{
		SslVpnClientCertId: tea.String(resourceID),
		RegionId:           tea.String(region),
	}

	_, err := s.Client.DeleteSslVpnClientCert(request)
	return err
}
