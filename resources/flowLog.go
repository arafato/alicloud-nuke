package resources

import (
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("flowLog", CollectFlowLogs)
}

// FlowLog represents an Alibaba Cloud VPC Flow Log resource
type FlowLog struct {
	Client *vpc.Client
	Region string
}

// CollectFlowLogs discovers all Flow Logs in the specified region
func CollectFlowLogs(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateVPCClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allFlowLogs []*vpc.DescribeFlowLogsResponseBodyFlowLogsFlowLog
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &vpc.DescribeFlowLogsRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeFlowLogs(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.FlowLogs != nil && response.Body.FlowLogs.FlowLog != nil {
			allFlowLogs = append(allFlowLogs, response.Body.FlowLogs.FlowLog...)
		}

		totalCount := ""
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		// TotalCount is a string in this API
		if totalCount == "" || int32(len(allFlowLogs)) >= int32(len(totalCount)) {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, flowLog := range allFlowLogs {
		flowLogID := ""
		if flowLog.FlowLogId != nil {
			flowLogID = *flowLog.FlowLogId
		}

		flowLogName := ""
		if flowLog.FlowLogName != nil {
			flowLogName = *flowLog.FlowLogName
		}
		if flowLogName == "" {
			flowLogName = flowLogID
		}

		res := types.Resource{
			Removable:    FlowLog{Client: client, Region: region},
			Region:       region,
			ResourceID:   flowLogID,
			ResourceName: flowLogName,
			ProductName:  "FlowLog",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Flow Log
func (f FlowLog) Remove(region string, resourceID string, resourceName string) error {
	request := &vpc.DeleteFlowLogRequest{
		FlowLogId: tea.String(resourceID),
		RegionId:  tea.String(region),
	}

	_, err := f.Client.DeleteFlowLog(request)
	return err
}
