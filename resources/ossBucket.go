package resources

import (
	"context"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("ossBucket", CollectOSSBuckets)
}

// OSSBucket represents an Alibaba Cloud OSS Bucket resource
type OSSBucket struct {
	Client *oss.Client
	Region string
}

// CollectOSSBuckets discovers all OSS buckets
// Note: OSS buckets are global but have a region, we filter by region
func CollectOSSBuckets(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateOSSClient(creds, region)
	if err != nil {
		return nil, err
	}

	// List all buckets
	request := &oss.ListBucketsRequest{}
	response, err := client.ListBuckets(context.Background(), request)
	if err != nil {
		return nil, err
	}

	var allResources types.Resources
	if response.Buckets != nil {
		for _, bucket := range response.Buckets {
			bucketName := ""
			if bucket.Name != nil {
				bucketName = *bucket.Name
			}

			// Filter by region
			bucketRegion := ""
			if bucket.Location != nil {
				// Location is like "oss-cn-hangzhou", extract region
				loc := *bucket.Location
				if len(loc) > 4 && loc[:4] == "oss-" {
					bucketRegion = loc[4:]
				} else {
					bucketRegion = loc
				}
			}

			// Only include buckets in this region
			if bucketRegion != region {
				continue
			}

			res := types.Resource{
				Removable:    OSSBucket{Client: client, Region: region},
				Region:       region,
				ResourceID:   bucketName,
				ResourceName: bucketName,
				ProductName:  "OSSBucket",
			}
			allResources = append(allResources, &res)
		}
	}

	return allResources, nil
}

// Remove deletes the OSS bucket (must be empty first)
func (o OSSBucket) Remove(region string, resourceID string, resourceName string) error {
	ctx := context.Background()

	// First delete all objects in the bucket
	paginator := o.Client.NewListObjectsV2Paginator(&oss.ListObjectsV2Request{
		Bucket: oss.Ptr(resourceID),
	})

	for paginator.HasNext() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		if len(page.Contents) > 0 {
			var objects []oss.DeleteObject
			for _, obj := range page.Contents {
				objects = append(objects, oss.DeleteObject{Key: obj.Key})
			}

			_, err = o.Client.DeleteMultipleObjects(ctx, &oss.DeleteMultipleObjectsRequest{
				Bucket:  oss.Ptr(resourceID),
				Objects: objects,
			})
			if err != nil {
				return err
			}
		}
	}

	// Delete all versions if versioning is enabled
	versionPaginator := o.Client.NewListObjectVersionsPaginator(&oss.ListObjectVersionsRequest{
		Bucket: oss.Ptr(resourceID),
	})

	for versionPaginator.HasNext() {
		page, err := versionPaginator.NextPage(ctx)
		if err != nil {
			// Versioning may not be enabled, continue
			break
		}

		// Delete object versions
		for _, version := range page.ObjectVersions {
			_, err = o.Client.DeleteObject(ctx, &oss.DeleteObjectRequest{
				Bucket:    oss.Ptr(resourceID),
				Key:       version.Key,
				VersionId: version.VersionId,
			})
			if err != nil {
				return err
			}
		}

		// Delete delete markers
		for _, marker := range page.ObjectDeleteMarkers {
			_, err = o.Client.DeleteObject(ctx, &oss.DeleteObjectRequest{
				Bucket:    oss.Ptr(resourceID),
				Key:       marker.Key,
				VersionId: marker.VersionId,
			})
			if err != nil {
				return err
			}
		}
	}

	// Now delete the bucket
	_, err := o.Client.DeleteBucket(ctx, &oss.DeleteBucketRequest{
		Bucket: oss.Ptr(resourceID),
	})
	return err
}
