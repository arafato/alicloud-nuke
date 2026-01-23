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
	"github.com/arafato/ali-nuke/utils"
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
	// "InvalidApi.NotFound" = API not available
	// "UnauthorizedRegion" = region not authorized for the service
	// "Forbidden." = permission denied for specific service features (e.g., Forbidden.HaVip)
	// "network is unreachable" = region not accessible from current network (e.g., geo-blocked)
	return strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "InvalidRegionId") ||
		strings.Contains(errStr, "InvalidApi.NotFound") ||
		strings.Contains(errStr, "UnauthorizedRegion") ||
		strings.Contains(errStr, "Forbidden.") ||
		strings.Contains(errStr, "network is unreachable")
}

// isTransientError returns true if the error is a transient network error that can be retried
func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// EOF = connection dropped
	// connection reset = server closed connection
	// timeout = request timed out
	// Note: Throttling errors are NOT retried here as they require longer backoffs
	return err == io.EOF ||
		strings.Contains(errStr, "EOF") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "timeout")
}

// isThrottlingError returns true if the error is a rate limiting error
func isThrottlingError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Throttling")
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

		// Don't retry non-transient errors (including throttling, service unavailable, etc.)
		if !isTransientError(err) {
			return nil, err
		}

		// Wait before retry (exponential backoff: 1s, 2s, 4s)
		if attempt < maxRetries {
			backoff := time.Duration(1<<attempt) * time.Second
			time.Sleep(backoff)
		}
	}
	return nil, lastErr
}

// ProcessCollection collects resources from all registered collectors across all specified regions
func ProcessCollection(creds *types.Credentials, regions []string, logger *utils.ScanLogger) types.Resources {
	var resourceCollectionChan = make(chan *types.Resource, 100)
	var allResources types.Resources
	g := new(errgroup.Group)
	// Limit concurrent API calls to avoid rate limiting
	g.SetLimit(20)

	for collectorName, collector := range collectors {
		for _, region := range regions {
			c := collector
			r := region
			cn := collectorName
			g.Go(func() error {
				resources, err := collectWithRetry(c, creds, r, 3)
				if err != nil {
					// Log but continue for "service not available in region" errors
					if isServiceUnavailableError(err) {
						logger.LogWarning("Service unavailable for %s in region %s: %v", cn, r, err)
						return nil
					}
					// Log but continue for throttling errors (we've already retried)
					if isThrottlingError(err) {
						logger.LogWarning("Throttling error for %s in region %s: %v", cn, r, err)
						return nil
					}
					// Log other errors to file but continue with other regions/collectors
					logger.LogWarning("Error collecting %s from region %s: %v", cn, r, err)
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
		logger.LogError("Fatal error during collection: %v", collectedErr)
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
