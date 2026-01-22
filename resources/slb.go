package resources

import (
	slb "github.com/alibabacloud-go/slb-20140515/v4/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("slb", CollectSLBInstances)
}

// SLB represents an Alibaba Cloud Classic Load Balancer (SLB) resource
type SLB struct {
	Client *slb.Client
	Region string
}

// CollectSLBInstances discovers all SLB instances in the specified region
func CollectSLBInstances(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateSLBClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allSLBs []*slb.DescribeLoadBalancersResponseBodyLoadBalancersLoadBalancer
	pageNumber := int32(1)
	pageSize := int32(100)

	for {
		request := &slb.DescribeLoadBalancersRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeLoadBalancers(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.LoadBalancers != nil && response.Body.LoadBalancers.LoadBalancer != nil {
			allSLBs = append(allSLBs, response.Body.LoadBalancers.LoadBalancer...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allSLBs)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, lb := range allSLBs {
		lbID := ""
		if lb.LoadBalancerId != nil {
			lbID = *lb.LoadBalancerId
		}

		lbName := ""
		if lb.LoadBalancerName != nil {
			lbName = *lb.LoadBalancerName
		}
		if lbName == "" {
			lbName = lbID
		}

		res := types.Resource{
			Removable:    SLB{Client: client, Region: region},
			Region:       region,
			ResourceID:   lbID,
			ResourceName: lbName,
			ProductName:  "SLB",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the SLB instance
func (s SLB) Remove(region string, resourceID string, resourceName string) error {
	request := &slb.DeleteLoadBalancerRequest{
		LoadBalancerId: tea.String(resourceID),
		RegionId:       tea.String(region),
	}

	_, err := s.Client.DeleteLoadBalancer(request)
	return err
}
