package resources

import (
	ess "github.com/alibabacloud-go/ess-20220222/v2/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("scalingConfiguration", CollectScalingConfigurations)
}

// ScalingConfiguration represents an Alibaba Cloud Auto Scaling Configuration resource
type ScalingConfiguration struct {
	Client *ess.Client
	Region string
}

// CollectScalingConfigurations discovers all Scaling Configurations in the specified region
func CollectScalingConfigurations(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateESSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allConfigs []*ess.DescribeScalingConfigurationsResponseBodyScalingConfigurations
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &ess.DescribeScalingConfigurationsRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeScalingConfigurations(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.ScalingConfigurations != nil {
			allConfigs = append(allConfigs, response.Body.ScalingConfigurations...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allConfigs)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, config := range allConfigs {
		configID := ""
		if config.ScalingConfigurationId != nil {
			configID = *config.ScalingConfigurationId
		}

		configName := ""
		if config.ScalingConfigurationName != nil {
			configName = *config.ScalingConfigurationName
		}
		if configName == "" {
			configName = configID
		}

		res := types.Resource{
			Removable:    ScalingConfiguration{Client: client, Region: region},
			Region:       region,
			ResourceID:   configID,
			ResourceName: configName,
			ProductName:  "ScalingConfiguration",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Scaling Configuration
func (s ScalingConfiguration) Remove(region string, resourceID string, resourceName string) error {
	request := &ess.DeleteScalingConfigurationRequest{
		ScalingConfigurationId: tea.String(resourceID),
	}

	_, err := s.Client.DeleteScalingConfiguration(request)
	return err
}
