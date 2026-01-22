package infrastructure

import (
	"github.com/arafato/ali-nuke/config"
	"github.com/arafato/ali-nuke/types"
)

func FilterCollection(resources types.Resources, config *config.Config) {
	// Resource type filter
	resourceTypeFilterSet := make(map[string]struct{})
	for _, filter := range config.ResourceTypes.Excludes {
		resourceTypeFilterSet[filter] = struct{}{}
	}

	// Region filter
	regionFilterSet := make(map[string]struct{})
	for _, region := range config.Regions.Excludes {
		regionFilterSet[region] = struct{}{}
	}

	// Resource ID filter
	resourceIDLookup := make(map[string]map[string]struct{})
	for _, resourceIDFilter := range config.ResourceIDs.Excludes {
		if _, ok := resourceIDLookup[resourceIDFilter.ResourceType]; !ok {
			resourceIDLookup[resourceIDFilter.ResourceType] = make(map[string]struct{})
		}
		resourceIDLookup[resourceIDFilter.ResourceType][resourceIDFilter.ID] = struct{}{}
	}

	for _, resource := range resources {
		if resource.State() == types.Hidden {
			continue
		}

		// Filter by resource type
		if _, ok := resourceTypeFilterSet[resource.ProductName]; ok {
			resource.SetState(types.Filtered)
			continue
		}

		// Filter by region
		if _, ok := regionFilterSet[resource.Region]; ok {
			resource.SetState(types.Filtered)
			continue
		}

		// Filter by resource ID or name
		if idSet, ok := resourceIDLookup[resource.ProductName]; ok {
			if _, ok := idSet[resource.ResourceID]; ok {
				resource.SetState(types.Filtered)
				continue
			}
			if _, ok := idSet[resource.ResourceName]; ok {
				resource.SetState(types.Filtered)
				continue
			}
		}

		resource.SetState(types.Ready)
	}
}
