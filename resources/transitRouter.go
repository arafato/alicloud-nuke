package resources

import (
	cbn "github.com/alibabacloud-go/cbn-20170912/v2/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("transitRouter", CollectTransitRouters)
}

// TransitRouter represents an Alibaba Cloud CEN Transit Router resource
type TransitRouter struct {
	Client *cbn.Client
	Region string
}

// CollectTransitRouters discovers all Transit Routers in the specified region
func CollectTransitRouters(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateCENClient(creds, region)
	if err != nil {
		return nil, err
	}

	// First, get all CEN instances to query their transit routers
	cenIDs, err := listAllCENIDs(client)
	if err != nil {
		return nil, err
	}

	var allResources types.Resources

	// For each CEN, list transit routers in the specified region
	for _, cenID := range cenIDs {
		transitRouters, err := listTransitRoutersForCEN(client, cenID, region)
		if err != nil {
			// Skip if there's an error for this CEN (e.g., permission issues)
			continue
		}

		for _, tr := range transitRouters {
			trID := ""
			if tr.TransitRouterId != nil {
				trID = *tr.TransitRouterId
			}

			trName := ""
			if tr.TransitRouterName != nil {
				trName = *tr.TransitRouterName
			}
			if trName == "" {
				trName = trID
			}

			res := types.Resource{
				Removable:    TransitRouter{Client: client, Region: region},
				Region:       region,
				ResourceID:   trID,
				ResourceName: trName,
				ProductName:  "TransitRouter",
			}
			allResources = append(allResources, &res)
		}
	}

	return allResources, nil
}

// listAllCENIDs returns all CEN instance IDs
func listAllCENIDs(client *cbn.Client) ([]string, error) {
	var cenIDs []string
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &cbn.DescribeCensRequest{
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeCens(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.Cens != nil && response.Body.Cens.Cen != nil {
			for _, cen := range response.Body.Cens.Cen {
				if cen.CenId != nil {
					cenIDs = append(cenIDs, *cen.CenId)
				}
			}
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(cenIDs)) >= totalCount {
			break
		}
		pageNumber++
	}

	return cenIDs, nil
}

// listTransitRoutersForCEN returns all transit routers for a CEN in the specified region
func listTransitRoutersForCEN(client *cbn.Client, cenID string, region string) ([]*cbn.ListTransitRoutersResponseBodyTransitRouters, error) {
	var allTransitRouters []*cbn.ListTransitRoutersResponseBodyTransitRouters
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &cbn.ListTransitRoutersRequest{
			CenId:      tea.String(cenID),
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.ListTransitRouters(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.TransitRouters != nil {
			allTransitRouters = append(allTransitRouters, response.Body.TransitRouters...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allTransitRouters)) >= totalCount {
			break
		}
		pageNumber++
	}

	return allTransitRouters, nil
}

// Remove deletes the Transit Router
func (t TransitRouter) Remove(region string, resourceID string, resourceName string) error {
	request := &cbn.DeleteTransitRouterRequest{
		TransitRouterId: tea.String(resourceID),
	}

	_, err := t.Client.DeleteTransitRouter(request)
	return err
}
