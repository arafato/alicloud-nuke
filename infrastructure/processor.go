package infrastructure

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/arafato/ali-nuke/types"
)

const (
	maxWaves     = 60               // Max number of retry waves
	waveInterval = 10 * time.Second // Time between waves
	maxTotalTime = 10 * time.Minute // Total timeout
)

// RemoveCollection attempts to remove all resources using a wave-based approach.
// Resources that fail with retriable errors (e.g., DependencyViolation) are retried
// in subsequent waves until they succeed, permanently fail, or timeout is reached.
func RemoveCollection(ctx context.Context, resources types.Resources) error {
	startTime := time.Now()

	for wave := 1; wave <= maxWaves; wave++ {
		// Check timeout
		if time.Since(startTime) > maxTotalTime {
			markPendingAsFailed(resources)
			break
		}

		// Count resources to process this wave (Ready or PendingRetry)
		toProcess := countProcessable(resources)
		if toProcess == 0 {
			break
		}

		// Reset PendingRetry → Ready just before processing
		resetPendingToReady(resources)

		// Run parallel deletion for this wave
		runDeletionWave(ctx, resources)

		// Check if any resources are pending retry
		pendingCount := resources.NumOf(types.PendingRetry)
		if pendingCount == 0 {
			break
		}

		// Wait before next wave (resources stay in PendingRetry state during wait)
		if wave < maxWaves && time.Since(startTime) < maxTotalTime {
			fmt.Printf("\nWave %d: %d resources need retry, waiting %v...\n", wave, pendingCount, waveInterval)
			select {
			case <-time.After(waveInterval):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}

// runDeletionWave processes all Ready resources in parallel
func runDeletionWave(ctx context.Context, resources types.Resources) {
	var wg sync.WaitGroup

	for _, resource := range resources {
		if resource.State() != types.Ready {
			continue
		}
		wg.Add(1)
		go func(r *types.Resource) {
			defer wg.Done()
			r.Remove(ctx)
		}(resource)
	}

	wg.Wait()
}

// countProcessable returns the number of resources that can be processed (Ready or PendingRetry)
func countProcessable(resources types.Resources) int {
	count := 0
	for _, r := range resources {
		if r.State() == types.Ready || r.State() == types.PendingRetry {
			count++
		}
	}
	return count
}

// resetPendingToReady resets PendingRetry resources back to Ready for the next wave
func resetPendingToReady(resources types.Resources) {
	for _, r := range resources {
		if r.State() == types.PendingRetry {
			r.SetState(types.Ready)
		}
	}
}

// markPendingAsFailed marks all PendingRetry resources as Failed (used on timeout)
func markPendingAsFailed(resources types.Resources) {
	for _, r := range resources {
		if r.State() == types.PendingRetry {
			r.SetState(types.Failed)
		}
	}
}
