package resources

import (
	nas "github.com/alibabacloud-go/nas-20170626/v3/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("nasFileSystem", CollectNASFileSystems)
}

// NASFileSystem represents an Alibaba Cloud NAS File System resource
type NASFileSystem struct {
	Client *nas.Client
	Region string
}

// CollectNASFileSystems discovers all NAS File Systems in the specified region
func CollectNASFileSystems(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateNASClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allFileSystems []*nas.DescribeFileSystemsResponseBodyFileSystemsFileSystem
	pageNumber := int32(1)
	pageSize := int32(100)

	for {
		request := &nas.DescribeFileSystemsRequest{
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		}

		response, err := client.DescribeFileSystems(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.FileSystems != nil && response.Body.FileSystems.FileSystem != nil {
			allFileSystems = append(allFileSystems, response.Body.FileSystems.FileSystem...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allFileSystems)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, fs := range allFileSystems {
		fsID := ""
		if fs.FileSystemId != nil {
			fsID = *fs.FileSystemId
		}

		fsDesc := ""
		if fs.Description != nil {
			fsDesc = *fs.Description
		}

		// Use description as name, or fall back to ID
		displayName := fsDesc
		if displayName == "" {
			displayName = fsID
		}

		res := types.Resource{
			Removable:    NASFileSystem{Client: client, Region: region},
			Region:       region,
			ResourceID:   fsID,
			ResourceName: displayName,
			ProductName:  "NASFileSystem",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the NAS File System
func (fs NASFileSystem) Remove(region string, resourceID string, resourceName string) error {
	request := &nas.DeleteFileSystemRequest{
		FileSystemId: tea.String(resourceID),
	}

	_, err := fs.Client.DeleteFileSystem(request)
	return err
}
