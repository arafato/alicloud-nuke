package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("snatEntry", CollectSnatEntries)
}

// SnatEntry represents an Alibaba Cloud SNAT Entry resource
type SnatEntry struct {
	Client      *vpc.Client
	Region      string
	SnatTableId string
}

// CollectSnatEntries discovers all SNAT Entries in the specified region
func CollectSnatEntries(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	// First get all NAT Gateways to find SNAT tables
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

	// For each NAT Gateway, get SNAT entries
	for _, nat := range allNatGateways {
		if nat.SnatTableIds == nil || nat.SnatTableIds.SnatTableId == nil {
			continue
		}

		for _, snatTableId := range nat.SnatTableIds.SnatTableId {
			if snatTableId == nil {
				continue
			}

			snatResp, err := client.DescribeSnatTableEntries(&vpc.DescribeSnatTableEntriesRequest{
				RegionId:    tea.String(region),
				SnatTableId: snatTableId,
			})
			if err != nil {
				continue
			}

			if snatResp.Body != nil && snatResp.Body.SnatTableEntries != nil && snatResp.Body.SnatTableEntries.SnatTableEntry != nil {
				for _, entry := range snatResp.Body.SnatTableEntries.SnatTableEntry {
					entryID := ""
					if entry.SnatEntryId != nil {
						entryID = *entry.SnatEntryId
					}

					entryName := ""
					if entry.SnatEntryName != nil {
						entryName = *entry.SnatEntryName
					}
					if entryName == "" && entry.SnatIp != nil {
						entryName = *entry.SnatIp
					}
					if entryName == "" {
						entryName = entryID
					}

					res := types.Resource{
						Removable:    SnatEntry{Client: client, Region: region, SnatTableId: *snatTableId},
						Region:       region,
						ResourceID:   entryID,
						ResourceName: entryName,
						ProductName:  "SnatEntry",
					}
					allResources = append(allResources, &res)
				}
			}
		}
	}

	return allResources, nil
}

// Remove deletes the SNAT Entry
func (s SnatEntry) Remove(region string, resourceID string, resourceName string) error {
	request := &vpc.DeleteSnatEntryRequest{
		SnatTableId: tea.String(s.SnatTableId),
		SnatEntryId: tea.String(resourceID),
		RegionId:    tea.String(region),
	}

	_, err := s.Client.DeleteSnatEntry(request)
	return err
}
