package resources

import (
	dds "github.com/alibabacloud-go/dds-20151201/v4/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("mongodbInstance", CollectMongoDBInstances)
}

// MongoDBInstance represents an Alibaba Cloud MongoDB Instance resource
type MongoDBInstance struct {
	Client *dds.Client
	Region string
}

// CollectMongoDBInstances discovers all MongoDB instances in the specified region
func CollectMongoDBInstances(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateMongoDBClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allInstances []*dds.DescribeDBInstancesResponseBodyDBInstancesDBInstance
	pageNumber := int32(1)
	pageSize := int32(30)

	for {
		request := &dds.DescribeDBInstancesRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeDBInstances(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.DBInstances != nil && response.Body.DBInstances.DBInstance != nil {
			allInstances = append(allInstances, response.Body.DBInstances.DBInstance...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
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
			Removable:    MongoDBInstance{Client: client, Region: region},
			Region:       region,
			ResourceID:   instanceID,
			ResourceName: instanceName,
			ProductName:  "MongoDBInstance",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the MongoDB instance
func (m MongoDBInstance) Remove(region string, resourceID string, resourceName string) error {
	request := &dds.DeleteDBInstanceRequest{
		DBInstanceId: tea.String(resourceID),
	}

	_, err := m.Client.DeleteDBInstance(request)
	return err
}
