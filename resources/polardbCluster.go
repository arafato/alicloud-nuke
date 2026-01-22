package resources

import (
	polardb "github.com/alibabacloud-go/polardb-20170801/v6/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("polardbCluster", CollectPolarDBClusters)
}

// PolarDBCluster represents an Alibaba Cloud PolarDB Cluster resource
type PolarDBCluster struct {
	Client *polardb.Client
	Region string
}

// CollectPolarDBClusters discovers all PolarDB clusters in the specified region
func CollectPolarDBClusters(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreatePolarDBClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allClusters []*polardb.DescribeDBClustersResponseBodyItemsDBCluster
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &polardb.DescribeDBClustersRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeDBClusters(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.Items != nil && response.Body.Items.DBCluster != nil {
			allClusters = append(allClusters, response.Body.Items.DBCluster...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalRecordCount != nil {
			totalCount = *response.Body.TotalRecordCount
		}

		if int32(len(allClusters)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, cluster := range allClusters {
		clusterID := ""
		if cluster.DBClusterId != nil {
			clusterID = *cluster.DBClusterId
		}

		clusterName := ""
		if cluster.DBClusterDescription != nil {
			clusterName = *cluster.DBClusterDescription
		}
		if clusterName == "" {
			clusterName = clusterID
		}

		res := types.Resource{
			Removable:    PolarDBCluster{Client: client, Region: region},
			Region:       region,
			ResourceID:   clusterID,
			ResourceName: clusterName,
			ProductName:  "PolarDBCluster",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the PolarDB cluster
func (p PolarDBCluster) Remove(region string, resourceID string, resourceName string) error {
	request := &polardb.DeleteDBClusterRequest{
		DBClusterId: tea.String(resourceID),
	}

	_, err := p.Client.DeleteDBCluster(request)
	return err
}
