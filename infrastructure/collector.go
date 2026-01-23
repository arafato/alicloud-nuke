package infrastructure

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

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

// isServiceUnavailableError returns true if the error indicates the service
// is not available in the region (not a real error, just unsupported)
func isServiceUnavailableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// "no such host" = DNS lookup failed, service endpoint doesn't exist in region
	// "InvalidRegionId" = region not supported by the service
	// "InvalidAccessKeyId.Inactive" in specific regions = service not activated
	return strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "InvalidRegionId") ||
		strings.Contains(errStr, "InvalidApi.NotFound")
}

// isTransientError returns true if the error is a transient network error that can be retried
func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	// EOF = connection dropped
	// connection reset = server closed connection
	// timeout = request timed out
	return err == io.EOF ||
		strings.Contains(err.Error(), "EOF") ||
		strings.Contains(err.Error(), "connection reset") ||
		strings.Contains(err.Error(), "timeout")
}

// collectWithRetry attempts to collect resources with retries for transient errors
func collectWithRetry(collector types.ResourceCollector, creds *types.Credentials, region string, maxRetries int) (types.Resources, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resources, err := collector(creds, region)
		if err == nil {
			return resources, nil
		}
		lastErr = err

		// Don't retry non-transient errors
		if !isTransientError(err) {
			return nil, err
		}

		// Wait before retry (exponential backoff: 1s, 2s, 4s)
		if attempt < maxRetries {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}
	}
	return nil, lastErr
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
				resources, err := collectWithRetry(c, creds, r, 3)
				if err != nil {
					// Silently ignore "service not available in region" errors
					if isServiceUnavailableError(err) {
						return nil
					}
					// Log other errors but continue with other regions/collectors
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

// ListCollector returns an alphabetically sorted list of registered collector names.
func ListCollectors() []string {
	var collectorNames []string
	for name, _ := range collectors {
		collectorNames = append(collectorNames, name)
	}
	slices.Sort(collectorNames)
	return collectorNames
}
