package utils

import (
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/types"
)

// FetchAllRegions retrieves all available Alibaba Cloud regions using the ECS DescribeRegions API.
// We use cn-hangzhou as the bootstrap region to query the list of all available regions.
func FetchAllRegions(creds *types.Credentials) ([]string, error) {
	// Create a client pointing to a known region to fetch the region list
	config := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		RegionId:        tea.String("cn-hangzhou"), // Bootstrap region
	}

	client, err := ecs.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create ECS client for region discovery: %w", err)
	}

	request := &ecs.DescribeRegionsRequest{
		AcceptLanguage: tea.String("en-US"), // English region names
	}

	response, err := client.DescribeRegions(request)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch regions: %w", err)
	}

	var regions []string
	if response.Body != nil && response.Body.Regions != nil && response.Body.Regions.Region != nil {
		for _, region := range response.Body.Regions.Region {
			if region.RegionId != nil {
				regions = append(regions, *region.RegionId)
			}
		}
	}

	if len(regions) == 0 {
		return nil, fmt.Errorf("no regions returned from DescribeRegions API")
	}

	return regions, nil
}

// GetActiveRegions fetches all available regions and removes the excluded ones
func GetActiveRegions(creds *types.Credentials, excludes []string) ([]string, error) {
	allRegions, err := FetchAllRegions(creds)
	if err != nil {
		return nil, err
	}

	excludeSet := make(map[string]struct{})
	for _, e := range excludes {
		excludeSet[e] = struct{}{}
	}

	var active []string
	for _, r := range allRegions {
		if _, excluded := excludeSet[r]; !excluded {
			active = append(active, r)
		}
	}

	return active, nil
}
