package resources

import (
	rds "github.com/alibabacloud-go/rds-20140815/v4/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("rdsInstance", CollectRDSInstances)
}

// RDSInstance represents an Alibaba Cloud RDS Instance resource
type RDSInstance struct {
	Client *rds.Client
	Region string
}

// CollectRDSInstances discovers all RDS instances in the specified region
func CollectRDSInstances(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateRDSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allInstances []*rds.DescribeDBInstancesResponseBodyItemsDBInstance
	pageNumber := int32(1)
	pageSize := int32(100)

	for {
		request := &rds.DescribeDBInstancesRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeDBInstances(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.Items != nil && response.Body.Items.DBInstance != nil {
			allInstances = append(allInstances, response.Body.Items.DBInstance...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalRecordCount != nil {
			totalCount = *response.Body.TotalRecordCount
		}

		if int32(len(allInstances)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, instance := range allInstances {
		instanceID := ""
		if instance.DBInstanceId != nil {
			instanceID = *instance.DBInstanceId
		}

		instanceName := ""
		if instance.DBInstanceDescription != nil {
			instanceName = *instance.DBInstanceDescription
		}
		if instanceName == "" {
			instanceName = instanceID
		}

		res := types.Resource{
			Removable:    RDSInstance{Client: client, Region: region},
			Region:       region,
			ResourceID:   instanceID,
			ResourceName: instanceName,
			ProductName:  "RDSInstance",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the RDS instance
func (r RDSInstance) Remove(region string, resourceID string, resourceName string) error {
	// First release the instance (for pay-as-you-go instances)
	request := &rds.DeleteDBInstanceRequest{
		DBInstanceId:       tea.String(resourceID),
		ReleasedKeepPolicy: tea.String("None"), // Don't keep backups
	}

	_, err := r.Client.DeleteDBInstance(request)
	return err
}
