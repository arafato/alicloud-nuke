package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("vswitch", CollectVSwitches)
}

// VSwitch represents an Alibaba Cloud VSwitch resource
type VSwitch struct {
	Client *vpc.Client
	Region string
}

// CollectVSwitches discovers all VSwitches in the specified region
func CollectVSwitches(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allVSwitches []*vpc.DescribeVSwitchesResponseBodyVSwitchesVSwitch
	pageNumber := int32(1)
	pageSize := int32(50)

	// Paginate through all VSwitches
	for {
		request := &vpc.DescribeVSwitchesRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeVSwitches(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.VSwitches != nil && response.Body.VSwitches.VSwitch != nil {
			allVSwitches = append(allVSwitches, response.Body.VSwitches.VSwitch...)
		}

		// Check if we've retrieved all VSwitches
		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allVSwitches)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, vs := range allVSwitches {
		vswitchID := ""
		if vs.VSwitchId != nil {
			vswitchID = *vs.VSwitchId
		}

		vswitchName := ""
		if vs.VSwitchName != nil {
			vswitchName = *vs.VSwitchName
		}
		// If no name, use VSwitch ID as the name for display
		if vswitchName == "" {
			vswitchName = vswitchID
		}

		res := types.Resource{
			Removable:    VSwitch{Client: client, Region: region},
			Region:       region,
			ResourceID:   vswitchID,
			ResourceName: vswitchName,
			ProductName:  "VSwitch",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the VSwitch
func (vs VSwitch) Remove(region string, resourceID string, resourceName string) error {
	request := &vpc.DeleteVSwitchRequest{
		VSwitchId: tea.String(resourceID),
		RegionId:  tea.String(region),
	}

	_, err := vs.Client.DeleteVSwitch(request)
	return err
}
