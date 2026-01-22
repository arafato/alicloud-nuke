package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("eip", CollectEIPs)
}

// EIP represents an Alibaba Cloud Elastic IP Address resource
type EIP struct {
	Client *vpc.Client
	Region string
}

// CollectEIPs discovers all Elastic IP Addresses in the specified region
func CollectEIPs(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allEIPs []*vpc.DescribeEipAddressesResponseBodyEipAddressesEipAddress
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &vpc.DescribeEipAddressesRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeEipAddresses(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.EipAddresses != nil && response.Body.EipAddresses.EipAddress != nil {
			allEIPs = append(allEIPs, response.Body.EipAddresses.EipAddress...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allEIPs)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, eip := range allEIPs {
		eipID := ""
		if eip.AllocationId != nil {
			eipID = *eip.AllocationId
		}

		// Use IP address as display name
		eipName := ""
		if eip.Name != nil && *eip.Name != "" {
			eipName = *eip.Name
		} else if eip.IpAddress != nil {
			eipName = *eip.IpAddress
		} else {
			eipName = eipID
		}

		res := types.Resource{
			Removable:    EIP{Client: client, Region: region},
			Region:       region,
			ResourceID:   eipID,
			ResourceName: eipName,
			ProductName:  "EIP",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Elastic IP Address
func (e EIP) Remove(region string, resourceID string, resourceName string) error {
	// First try to unassociate if attached
	unassociateReq := &vpc.UnassociateEipAddressRequest{
		AllocationId: tea.String(resourceID),
		RegionId:     tea.String(region),
		Force:        tea.Bool(true),
	}
	// Ignore unassociate errors - it may not be associated
	e.Client.UnassociateEipAddress(unassociateReq)

	// Release the EIP
	request := &vpc.ReleaseEipAddressRequest{
		AllocationId: tea.String(resourceID),
		RegionId:     tea.String(region),
	}

	_, err := e.Client.ReleaseEipAddress(request)
	return err
}
