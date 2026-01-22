package resources

import (
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("launchTemplate", CollectLaunchTemplates)
}

// LaunchTemplate represents an Alibaba Cloud ECS Launch Template resource
type LaunchTemplate struct {
	Client *ecs.Client
	Region string
}

// CollectLaunchTemplates discovers all Launch Templates in the specified region
func CollectLaunchTemplates(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateECSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allTemplates []*ecs.DescribeLaunchTemplatesResponseBodyLaunchTemplateSetsLaunchTemplateSet
	pageNumber := int32(1)
	pageSize := int32(50)

	for {
		request := &ecs.DescribeLaunchTemplatesRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeLaunchTemplates(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.LaunchTemplateSets != nil && response.Body.LaunchTemplateSets.LaunchTemplateSet != nil {
			allTemplates = append(allTemplates, response.Body.LaunchTemplateSets.LaunchTemplateSet...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allTemplates)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, template := range allTemplates {
		templateID := ""
		if template.LaunchTemplateId != nil {
			templateID = *template.LaunchTemplateId
		}

		templateName := ""
		if template.LaunchTemplateName != nil {
			templateName = *template.LaunchTemplateName
		}
		if templateName == "" {
			templateName = templateID
		}

		res := types.Resource{
			Removable:    LaunchTemplate{Client: client, Region: region},
			Region:       region,
			ResourceID:   templateID,
			ResourceName: templateName,
			ProductName:  "LaunchTemplate",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Launch Template
func (l LaunchTemplate) Remove(region string, resourceID string, resourceName string) error {
	request := &ecs.DeleteLaunchTemplateRequest{
		RegionId:         tea.String(region),
		LaunchTemplateId: tea.String(resourceID),
	}

	_, err := l.Client.DeleteLaunchTemplate(request)
	return err
}
