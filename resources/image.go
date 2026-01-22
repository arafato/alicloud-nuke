package resources

import (
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("image", CollectImages)
}

// Image represents an Alibaba Cloud ECS Custom Image resource
type Image struct {
	Client *ecs.Client
	Region string
}

// CollectImages discovers all custom Images in the specified region
func CollectImages(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateECSClient(creds, region)
	if err != nil {
		return nil, err
	}

	var allImages []*ecs.DescribeImagesResponseBodyImagesImage
	pageNumber := int32(1)
	pageSize := int32(100)

	for {
		request := &ecs.DescribeImagesRequest{
			RegionId:        tea.String(region),
			PageNumber:      tea.Int32(pageNumber),
			PageSize:        tea.Int32(pageSize),
			ImageOwnerAlias: tea.String("self"), // Only get custom images, not public/marketplace
		}

		response, err := client.DescribeImages(request)
		if err != nil {
			return nil, err
		}

		if response.Body != nil && response.Body.Images != nil && response.Body.Images.Image != nil {
			allImages = append(allImages, response.Body.Images.Image...)
		}

		totalCount := int32(0)
		if response.Body != nil && response.Body.TotalCount != nil {
			totalCount = *response.Body.TotalCount
		}

		if int32(len(allImages)) >= totalCount {
			break
		}
		pageNumber++
	}

	var allResources types.Resources
	for _, image := range allImages {
		imageID := ""
		if image.ImageId != nil {
			imageID = *image.ImageId
		}

		imageName := ""
		if image.ImageName != nil {
			imageName = *image.ImageName
		}
		if imageName == "" {
			imageName = imageID
		}

		res := types.Resource{
			Removable:    Image{Client: client, Region: region},
			Region:       region,
			ResourceID:   imageID,
			ResourceName: imageName,
			ProductName:  "Image",
		}
		allResources = append(allResources, &res)
	}

	return allResources, nil
}

// Remove deletes the Image
func (i Image) Remove(region string, resourceID string, resourceName string) error {
	request := &ecs.DeleteImageRequest{
		ImageId:  tea.String(resourceID),
		RegionId: tea.String(region),
		Force:    tea.Bool(true), // Force delete even if used by instances
	}

	_, err := i.Client.DeleteImage(request)
	return err
}
