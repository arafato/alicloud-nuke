package resources

import (
	nlb "github.com/alibabacloud-go/nlb-20220430/v2/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("nlb", CollectNLBInstances)
}

// NLB represents an Alibaba Cloud Network Load Balancer resource
type NLB struct {
	Client *nlb.Client
	Region string
}

// CollectNLBInstances discovers all NLB instances in the specified region
func CollectNLBInstances(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateNLBClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allNLBs []*nlb.ListLoadBalancersResponseBodyLoadBalancers
	nextToken := ""

	for {
		request := &nlb.ListLoadBalancersRequest{
			RegionId:   tea.String(region),
			MaxResults: tea.Int32(100),
		}
		if nextToken != "" {
			request.NextToken = tea.String(nextToken)
		}

		response, err := client.ListLoadBalancers(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.LoadBalancers != nil {
			allNLBs = append(allNLBs, response.Body.LoadBalancers...)
		}

		if response.Body == nil || response.Body.NextToken == nil || *response.Body.NextToken == "" {
			break
		}
		nextToken = *response.Body.NextToken
	}

	var allResources types.Resources
	for _, lb := range allNLBs {
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
			Removable:    NLB{Client: client, Region: region},
			Region:       region,
			ResourceID:   lbID,
			ResourceName: lbName,
			ProductName:  "NLB",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the NLB instance
func (n NLB) Remove(region string, resourceID string, resourceName string) error {
	request := &nlb.DeleteLoadBalancerRequest{
		LoadBalancerId: tea.String(resourceID),
		RegionId:       tea.String(region),
	}

	_, err := n.Client.DeleteLoadBalancer(request)
	return err
}
