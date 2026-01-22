package resources

import (
	alb "github.com/alibabacloud-go/alb-20200616/v2/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("alb", CollectALBInstances)
}

// ALB represents an Alibaba Cloud Application Load Balancer resource
type ALB struct {
	Client *alb.Client
	Region string
}

// CollectALBInstances discovers all ALB instances in the specified region
func CollectALBInstances(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateALBClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allALBs []*alb.ListLoadBalancersResponseBodyLoadBalancers
	nextToken := ""

	for {
		request := &alb.ListLoadBalancersRequest{
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
			allALBs = append(allALBs, response.Body.LoadBalancers...)
		}

		if response.Body == nil || response.Body.NextToken == nil || *response.Body.NextToken == "" {
			break
		}
		nextToken = *response.Body.NextToken
	}

	var allResources types.Resources
	for _, lb := range allALBs {
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
			Removable:    ALB{Client: client, Region: region},
			Region:       region,
			ResourceID:   lbID,
			ResourceName: lbName,
			ProductName:  "ALB",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the ALB instance
func (a ALB) Remove(region string, resourceID string, resourceName string) error {
	request := &alb.DeleteLoadBalancerRequest{
		LoadBalancerId: tea.String(resourceID),
	}

	_, err := a.Client.DeleteLoadBalancer(request)
	return err
}
