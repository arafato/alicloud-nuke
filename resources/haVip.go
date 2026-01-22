package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("haVip", CollectHaVips)
}

// HaVip represents an Alibaba Cloud High Availability Virtual IP resource
type HaVip struct {
	Client *vpc.Client
	Region string
}

// CollectHaVips discovers all HA VIPs in the specified region
func CollectHaVips(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allHaVips []*vpc.DescribeHaVipsResponseBodyHaVipsHaVip
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &vpc.DescribeHaVipsRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeHaVips(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.HaVips != nil && response.Body.HaVips.HaVip != nil {
			allHaVips = append(allHaVips, response.Body.HaVips.HaVip...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allHaVips)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, havip := range allHaVips {
		havipID := ""
		if havip.HaVipId != nil {
			havipID = *havip.HaVipId
		}

		havipName := ""
		if havip.Name != nil {
			havipName = *havip.Name
		}
		if havipName == "" && havip.IpAddress != nil {
			havipName = *havip.IpAddress
		}
		if havipName == "" {
			havipName = havipID
		}

		res := types.Resource{
			Removable:    HaVip{Client: client, Region: region},
			Region:       region,
			ResourceID:   havipID,
			ResourceName: havipName,
			ProductName:  "HaVip",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the HA VIP
func (h HaVip) Remove(region string, resourceID string, resourceName string) error {
	request := &vpc.DeleteHaVipRequest{
		HaVipId:  tea.String(resourceID),
		RegionId: tea.String(region),
	}

	_, err := h.Client.DeleteHaVip(request)
	return err
}
