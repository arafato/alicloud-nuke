package infrastructure

import (
	"fmt"
	"os"

	"golang.org/x/sync/errgroup"

	"github.com/arafato/ali-nuke/types"
)

var collectors = make(map[string]types.ResourceCollector)

func RegisterCollector(name string, collector types.ResourceCollector) {
	if _, exists := collectors[name]; exists {
		panic(fmt.Errorf("handler %s already registered", name))
	}
	collectors[name] = collector
}

// ProcessCollection collects resources from all registered collectors across all specified regions
func ProcessCollection(creds *types.Credentials, regions []string) types.Resources {
	var resourceCollectionChan = make(chan *types.Resource, 100)
	var allResources types.Resources
	g := new(errgroup.Group)

	for _, collector := range collectors {
		for _, region := range regions {
			c := collector
			r := region
			g.Go(func() error {
				resources, err := c(creds, r)
				if err != nil {
					// Log error but continue with other regions/collectors
					fmt.Fprintf(os.Stderr, "Warning: Error collecting from region %s: %v\n", r, err)
					return nil
				}
				for _, resource := range resources {
					resourceCollectionChan <- resource
				}
				return nil
			})
		}
	}

	var collectedErr error
	go func() {
		collectedErr = g.Wait()
		close(resourceCollectionChan)
	}()

	for resource := range resourceCollectionChan {
		allResources = append(allResources, resource)
	}

	if collectedErr != nil {
		fmt.Println("Error during collection, aborting:\n", collectedErr)
		os.Exit(1)
	}

	return allResources
}
