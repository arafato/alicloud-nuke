package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("vpc", CollectVPCs)
}

// VPC represents an Alibaba Cloud VPC resource
type VPC struct {
	Client *vpc.Client
	Region string
}

// CollectVPCs discovers all VPCs in the specified region
func CollectVPCs(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allVPCs []*vpc.DescribeVpcsResponseBodyVpcsVpc
	pageNumber := int32(1)
	pageSize := int32(50)

	// Paginate through all VPCs
	for {
		request := &vpc.DescribeVpcsRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeVpcs(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.Vpcs != nil && response.Body.Vpcs.Vpc != nil {
			allVPCs = append(allVPCs, response.Body.Vpcs.Vpc...)
		}

		// Check if we've retrieved all VPCs
		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allVPCs)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, v := range allVPCs {
		vpcID := ""
		if v.VpcId != nil {
			vpcID = *v.VpcId
		}

		vpcName := ""
		if v.VpcName != nil {
			vpcName = *v.VpcName
		}
		// If no name, use VPC ID as the name for display
		if vpcName == "" {
			vpcName = vpcID
		}

		res := types.Resource{
			Removable:    VPC{Client: client, Region: region},
			Region:       region,
			ResourceID:   vpcID,
			ResourceName: vpcName,
			ProductName:  "VPC",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the VPC
func (v VPC) Remove(region string, resourceID string, resourceName string) error {
	request := &vpc.DeleteVpcRequest{
		VpcId:    tea.String(resourceID),
		RegionId: tea.String(region),
	}

	_, err := v.Client.DeleteVpc(request)
	return err
}
