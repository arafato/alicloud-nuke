package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("forwardEntry", CollectForwardEntries)
}

// ForwardEntry represents an Alibaba Cloud Forward Entry (DNAT) resource
type ForwardEntry struct {
	Client         *vpc.Client
	Region         string
	ForwardTableId string
}

// CollectForwardEntries discovers all Forward Entries (DNAT) in the specified region
func CollectForwardEntries(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	// First get all NAT Gateways to find Forward tables
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

	// For each NAT Gateway, get Forward entries
	for _, nat := range allNatGateways {
		if nat.ForwardTableIds == nil || nat.ForwardTableIds.ForwardTableId == nil {
			continue
		}

		for _, forwardTableId := range nat.ForwardTableIds.ForwardTableId {
			if forwardTableId == nil {
				continue
			}

			fwdResp, err := client.DescribeForwardTableEntries(&vpc.DescribeForwardTableEntriesRequest{
				RegionId:       tea.String(region),
				ForwardTableId: forwardTableId,
			})
			if err != nil {
				continue
			}

			if fwdResp.Body != nil && fwdResp.Body.ForwardTableEntries != nil && fwdResp.Body.ForwardTableEntries.ForwardTableEntry != nil {
				for _, entry := range fwdResp.Body.ForwardTableEntries.ForwardTableEntry {
					entryID := ""
					if entry.ForwardEntryId != nil {
						entryID = *entry.ForwardEntryId
					}

					// Build a descriptive name: ExternalIP:ExternalPort -> InternalIP:InternalPort
					entryName := ""
					if entry.ForwardEntryName != nil && *entry.ForwardEntryName != "" {
						entryName = *entry.ForwardEntryName
					} else {
						extIP := tea.StringValue(entry.ExternalIp)
						extPort := tea.StringValue(entry.ExternalPort)
						intIP := tea.StringValue(entry.InternalIp)
						intPort := tea.StringValue(entry.InternalPort)
						if extIP != "" && intIP != "" {
							entryName = extIP + ":" + extPort + "->" + intIP + ":" + intPort
						} else {
							entryName = entryID
						}
					}

					res := types.Resource{
						Removable:    ForwardEntry{Client: client, Region: region, ForwardTableId: *forwardTableId},
						Region:       region,
						ResourceID:   entryID,
						ResourceName: entryName,
						ProductName:  "ForwardEntry",
					}
					allResources = append(allResources, &res)
				}
			}
		}
	}

	return allResources, nil
}

// Remove deletes the Forward Entry
func (f ForwardEntry) Remove(region string, resourceID string, resourceName string) error {
	request := &vpc.DeleteForwardEntryRequest{
		ForwardTableId: tea.String(f.ForwardTableId),
		ForwardEntryId: tea.String(resourceID),
		RegionId:       tea.String(region),
	}

	_, err := f.Client.DeleteForwardEntry(request)
	return err
}
