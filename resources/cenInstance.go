package resources

import (
	cbn "github.com/alibabacloud-go/cbn-20170912/v2/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("cenInstance", CollectCENInstances)
}

// CENInstance represents an Alibaba Cloud Cloud Enterprise Network (CEN) instance resource
type CENInstance struct {
	Client *cbn.Client
	Region string
}

// CollectCENInstances discovers all CEN instances
// Note: CEN instances are global resources, we only collect them once using a fixed region
func CollectCENInstances(creds *types.Credentials, region string) (types.Resources, error) {
	// CEN is a global service, only collect from cn-hangzhou to avoid duplicates
	if region != "cn-hangzhou" {
		return nil, nil
	}

	client, err := utils.CreateCENClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allCens []*cbn.DescribeCensResponseBodyCensCen
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
			allCens = append(allCens, response.Body.Cens.Cen...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allCens)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, cen := range allCens {
		cenID := ""
		if cen.CenId != nil {
			cenID = *cen.CenId
		}

		cenName := ""
		if cen.Name != nil {
			cenName = *cen.Name
		}
		if cenName == "" {
			cenName = cenID
		}

		res := types.Resource{
			Removable:    CENInstance{Client: client, Region: region},
			Region:       "global", // CEN is a global resource
			ResourceID:   cenID,
			ResourceName: cenName,
			ProductName:  "CENInstance",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the CEN instance
func (c CENInstance) Remove(region string, resourceID string, resourceName string) error {
	request := &cbn.DeleteCenRequest{
		CenId: tea.String(resourceID),
	}

	_, err := c.Client.DeleteCen(request)
	return err
}
