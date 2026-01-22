package resources

import (
	nas "github.com/alibabacloud-go/nas-20170626/v3/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("nasMountTarget", CollectNASMountTargets)
}

// NASMountTarget represents an Alibaba Cloud NAS Mount Target resource
type NASMountTarget struct {
	Client       *nas.Client
	Region       string
	FileSystemID string // Required for deletion
}

// CollectNASMountTargets discovers all NAS Mount Targets in the specified region
func CollectNASMountTargets(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateNASClient(creds, region)
	if err != nil {
		return nil, err
	}

	// First, get all file systems
	var allFileSystems []*nas.DescribeFileSystemsResponseBodyFileSystemsFileSystem
	pageNumber := int32(1)
	pageSize := int32(100)

	for {
		fsRequest := &nas.DescribeFileSystemsRequest{
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		fsResponse, err := client.DescribeFileSystems(fsRequest)
		if err != nil {
			return nil, err
		}

		if fsResponse.Body != nil && fsResponse.Body.FileSystems != nil && fsResponse.Body.FileSystems.FileSystem != nil {
			allFileSystems = append(allFileSystems, fsResponse.Body.FileSystems.FileSystem...)
		}

		totalCount := int32(0)
		if fsResponse.Body != nil && fsResponse.Body.TotalCount != nil {
			totalCount = *fsResponse.Body.TotalCount
		}

		if int32(len(allFileSystems)) >= totalCount {
			break
		}
		pageNumber++
	}

	// For each file system, get its mount targets
	var allResources types.Resources
	for _, fs := range allFileSystems {
		fsID := ""
		if fs.FileSystemId != nil {
			fsID = *fs.FileSystemId
		}

		mtRequest := &nas.DescribeMountTargetsRequest{
			FileSystemId: tea.String(fsID),
		}

		mtResponse, err := client.DescribeMountTargets(mtRequest)
		if err != nil {
			// Log but continue with other file systems
			continue
		}

		if mtResponse.Body != nil && mtResponse.Body.MountTargets != nil && mtResponse.Body.MountTargets.MountTarget != nil {
			for _, mt := range mtResponse.Body.MountTargets.MountTarget {
				mtDomain := ""
				if mt.MountTargetDomain != nil {
					mtDomain = *mt.MountTargetDomain
				}

				// Display name shows mount target domain and parent file system
				displayName := mtDomain
				if displayName == "" {
					displayName = fsID + " (mount target)"
				}

				res := types.Resource{
					Removable:    NASMountTarget{Client: client, Region: region, FileSystemID: fsID},
					Region:       region,
					ResourceID:   mtDomain, // Mount target domain is the ID
					ResourceName: displayName,
					ProductName:  "NASMountTarget",
				}
				allResources = append(allResources, &res)
			}
		}
	}

	return allResources, nil
}

// Remove deletes the NAS Mount Target
func (mt NASMountTarget) Remove(region string, resourceID string, resourceName string) error {
	request := &nas.DeleteMountTargetRequest{
		FileSystemId:      tea.String(mt.FileSystemID),
		MountTargetDomain: tea.String(resourceID),
	}

	_, err := mt.Client.DeleteMountTarget(request)
	return err
}
