package resources

import (
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("disk", CollectDisks)
}

// Disk represents an Alibaba Cloud ECS Disk resource
type Disk struct {
	Client *ecs.Client
	Region string
}

// CollectDisks discovers all Disks in the specified region
func CollectDisks(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateECSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allDisks []*ecs.DescribeDisksResponseBodyDisksDisk
	pageNumber := int32(1)
	pageSize := int32(100)

	for {
		request := &ecs.DescribeDisksRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeDisks(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.Disks != nil && response.Body.Disks.Disk != nil {
			allDisks = append(allDisks, response.Body.Disks.Disk...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allDisks)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, disk := range allDisks {
		diskID := ""
		if disk.DiskId != nil {
			diskID = *disk.DiskId
		}

		diskName := ""
		if disk.DiskName != nil {
			diskName = *disk.DiskName
		}
		if diskName == "" {
			diskName = diskID
		}

		// Skip system disks - they are deleted with the instance
		diskType := ""
		if disk.Type != nil {
			diskType = *disk.Type
		}
		if diskType == "system" {
			continue
		}

		// Skip disks attached to instances (they'll be deleted with the instance or need to be detached first)
		// We only delete unattached data disks
		status := ""
		if disk.Status != nil {
			status = *disk.Status
		}

		res := types.Resource{
			Removable:    Disk{Client: client, Region: region},
			Region:       region,
			ResourceID:   diskID,
			ResourceName: diskName,
			ProductName:  "Disk",
		}

		// Hide attached disks - they need instance deletion first
		if status == "In_use" {
			res.SetState(types.Hidden)
		}

		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Disk
func (d Disk) Remove(region string, resourceID string, resourceName string) error {
	request := &ecs.DeleteDiskRequest{
		DiskId: tea.String(resourceID),
	}

	_, err := d.Client.DeleteDisk(request)
	return err
}
