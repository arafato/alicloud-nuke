package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("commonBandwidthPackage", CollectCommonBandwidthPackages)
}

// CommonBandwidthPackage represents an Alibaba Cloud Common Bandwidth Package resource
type CommonBandwidthPackage struct {
	Client *vpc.Client
	Region string
}

// CollectCommonBandwidthPackages discovers all Common Bandwidth Packages in the specified region
func CollectCommonBandwidthPackages(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allPackages []*vpc.DescribeCommonBandwidthPackagesResponseBodyCommonBandwidthPackagesCommonBandwidthPackage
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &vpc.DescribeCommonBandwidthPackagesRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeCommonBandwidthPackages(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.CommonBandwidthPackages != nil && response.Body.CommonBandwidthPackages.CommonBandwidthPackage != nil {
			allPackages = append(allPackages, response.Body.CommonBandwidthPackages.CommonBandwidthPackage...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allPackages)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, pkg := range allPackages {
		pkgID := ""
		if pkg.BandwidthPackageId != nil {
			pkgID = *pkg.BandwidthPackageId
		}

		pkgName := ""
		if pkg.Name != nil {
			pkgName = *pkg.Name
		}
		if pkgName == "" {
			pkgName = pkgID
		}

		res := types.Resource{
			Removable:    CommonBandwidthPackage{Client: client, Region: region},
			Region:       region,
			ResourceID:   pkgID,
			ResourceName: pkgName,
			ProductName:  "CommonBandwidthPackage",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Common Bandwidth Package
func (c CommonBandwidthPackage) Remove(region string, resourceID string, resourceName string) error {
	request := &vpc.DeleteCommonBandwidthPackageRequest{
		BandwidthPackageId: tea.String(resourceID),
		RegionId:           tea.String(region),
		Force:              tea.String("true"), // Force delete even if EIPs are associated
	}

	_, err := c.Client.DeleteCommonBandwidthPackage(request)
	return err
}
