package resources

import (
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("snapshot", CollectSnapshots)
}

// Snapshot represents an Alibaba Cloud ECS Snapshot resource
type Snapshot struct {
	Client *ecs.Client
	Region string
}

// CollectSnapshots discovers all Snapshots in the specified region
func CollectSnapshots(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateECSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allSnapshots []*ecs.DescribeSnapshotsResponseBodySnapshotsSnapshot
	pageNumber := int32(1)
	pageSize := int32(100)

	for {
		request := &ecs.DescribeSnapshotsRequest{
			RegionId:   tea.String(region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeSnapshots(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.Snapshots != nil && response.Body.Snapshots.Snapshot != nil {
			allSnapshots = append(allSnapshots, response.Body.Snapshots.Snapshot...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allSnapshots)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, snapshot := range allSnapshots {
		snapshotID := ""
		if snapshot.SnapshotId != nil {
			snapshotID = *snapshot.SnapshotId
		}

		snapshotName := ""
		if snapshot.SnapshotName != nil {
			snapshotName = *snapshot.SnapshotName
		}
		if snapshotName == "" {
			snapshotName = snapshotID
		}

		res := types.Resource{
			Removable:    Snapshot{Client: client, Region: region},
			Region:       region,
			ResourceID:   snapshotID,
			ResourceName: snapshotName,
			ProductName:  "Snapshot",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Snapshot
func (s Snapshot) Remove(region string, resourceID string, resourceName string) error {
	request := &ecs.DeleteSnapshotRequest{
		SnapshotId: tea.String(resourceID),
		Force:      tea.Bool(true), // Force delete even if used by custom images
	}

	_, err := s.Client.DeleteSnapshot(request)
	return err
}
