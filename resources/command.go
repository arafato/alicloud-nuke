package resources

import (
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("command", CollectCommands)
}

// Command represents an Alibaba Cloud ECS Cloud Assistant Command resource
type Command struct {
	Client *ecs.Client
	Region string
}

// CollectCommands discovers all Cloud Assistant Commands in the specified region
func CollectCommands(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateECSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allCommands []*ecs.DescribeCommandsResponseBodyCommandsCommand
	pageNumber := int64(1)
	pageSize := int64(50)

	for {
		request := &ecs.DescribeCommandsRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int64(pageNumber),
			PageSize:   tea.Int64(pageSize),
		}

		response, err := client.DescribeCommands(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.Commands != nil && response.Body.Commands.Command != nil {
			allCommands = append(allCommands, response.Body.Commands.Command...)
		}

		totalCount := int64(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int64(len(allCommands)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, cmd := range allCommands {
		cmdID := ""
		if cmd.CommandId != nil {
			cmdID = *cmd.CommandId
		}

		cmdName := ""
		if cmd.Name != nil {
			cmdName = *cmd.Name
		}
		if cmdName == "" {
			cmdName = cmdID
		}

		// Skip system commands (Provider != User)
		provider := ""
		if cmd.Provider != nil {
			provider = *cmd.Provider
		}
		if provider != "" && provider != "User" {
			continue
		}

		res := types.Resource{
			Removable:    Command{Client: client, Region: region},
			Region:       region,
			ResourceID:   cmdID,
			ResourceName: cmdName,
			ProductName:  "Command",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Command
func (c Command) Remove(region string, resourceID string, resourceName string) error {
	request := &ecs.DeleteCommandRequest{
		RegionId:  tea.String(region),
		CommandId: tea.String(resourceID),
	}

	_, err := c.Client.DeleteCommand(request)
	return err
}
