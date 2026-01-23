package resources

import (
	cs "github.com/alibabacloud-go/cs-20151215/v5/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("ackCluster", CollectACKClusters)
}

// ACKCluster represents an Alibaba Cloud Container Service for Kubernetes (ACK) cluster resource
type ACKCluster struct {
	Client *cs.Client
	Region string
}

// CollectACKClusters discovers all ACK clusters in the specified region
func CollectACKClusters(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateCSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allClusters []*cs.DescribeClustersV1ResponseBodyClusters
	pageNumber := int64(1)
	pageSize := int64(50)

	for {
		request := &cs.DescribeClustersV1Request{
			RegionId:   tea.String(region),
			PageNumber: tea.Int64(pageNumber),
			PageSize:   tea.Int64(pageSize),
		}

		response, err := client.DescribeClustersV1(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.Clusters != nil {
			allClusters = append(allClusters, response.Body.Clusters...)
		}

		totalCount := int64(0)
		if response.Body != nil && response.Body.PageInfo != nil && response.Body.PageInfo.TotalCount != nil {
			totalCount = int64(*response.Body.PageInfo.TotalCount)
		}

		if int64(len(allClusters)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, cluster := range allClusters {
		clusterID := ""
		if cluster.ClusterId != nil {
			clusterID = *cluster.ClusterId
		}

		clusterName := ""
		if cluster.Name != nil {
			clusterName = *cluster.Name
		}
		if clusterName == "" {
			clusterName = clusterID
		}

		res := types.Resource{
			Removable:    ACKCluster{Client: client, Region: region},
			Region:       region,
			ResourceID:   clusterID,
			ResourceName: clusterName,
			ProductName:  "ACKCluster",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the ACK cluster
func (a ACKCluster) Remove(region string, resourceID string, resourceName string) error {
	request := &cs.DeleteClusterRequest{
		// Do not retain resources - delete everything associated with the cluster
		RetainAllResources: tea.Bool(false),
	}

	_, err := a.Client.DeleteCluster(tea.String(resourceID), request)
	return err
}
