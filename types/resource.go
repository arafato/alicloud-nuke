package types

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v5"
)

type Removable interface {
	// Remove deletes the resource. Region is passed for resources that need regional context.
	Remove(region string, resourceID string, resourceName string) error
}

type Resource struct {
	Removable
	Region       string // Alibaba Cloud region ID (e.g., "cn-hangzhou")
	ResourceID   string
	ResourceName string
	ProductName  string
	state        atomic.Int32 // use State() and SetState() for thread-safe access
}

// ResourceCollector is a function that collects resources of a specific type in a given region
type ResourceCollector func(creds *Credentials, region string) (Resources, error)

type Resources []*Resource

//go:generate stringer -type=ResourceState
type ResourceState int32

const (
	// Ready is the default state (zero value) for new resources
	Ready ResourceState = iota
	Removing
	Deleted
	Failed
	Filtered
	Hidden
	PendingRetry // Failed with retriable error, will be retried in next wave
)

// State returns the current state of the resource (thread-safe)
func (r *Resource) State() ResourceState {
	return ResourceState(r.state.Load())
}

// SetState sets the state of the resource (thread-safe)
func (r *Resource) SetState(s ResourceState) {
	r.state.Store(int32(s))
}

// isPermanentError returns true for errors that should not be retried
func isPermanentError(errStr string) bool {
	permanentErrors := []string{
		"403",
		"Forbidden",
		"InvalidAccessKeyId",
		"InvalidResourceId.NotFound",
		"InvalidInstanceId.NotFound",
		"InvalidVpcId.NotFound",
		"InvalidVSwitchId.NotFound",
	}
	for _, e := range permanentErrors {
		if strings.Contains(errStr, e) {
			return true
		}
	}
	return false
}

// isRetriableError returns true for errors that can succeed after dependencies are resolved
func isRetriableError(errStr string) bool {
	retriableErrors := []string{
		"DependencyViolation",
		"IncorrectInstanceStatus",
		"OperationConflict",
		"ServiceUnavailable",
		"Throttling",
		"HasMountTarget", // NAS file system has mount targets
	}
	for _, e := range retriableErrors {
		if strings.Contains(errStr, e) {
			return true
		}
	}
	return false
}

// Remove attempts to delete the resource with retries for transient errors.
// Sets state to Deleted on success, Failed on permanent error, PendingRetry on retriable error.
func (r *Resource) Remove(ctx context.Context) error {
	r.SetState(Removing)

	// Configure backoff for quick retries within a wave (handles transient network issues)
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 1 * time.Second
	expBackoff.MaxInterval = 5 * time.Second

	var lastErr error
	operation := func() (struct{}, error) {
		err := r.Removable.Remove(r.Region, r.ResourceID, r.ResourceName)
		if err != nil {
			lastErr = err
			errStr := err.Error()

			// Permanent errors - stop retrying immediately
			if isPermanentError(errStr) {
				return struct{}{}, backoff.Permanent(err)
			}

			// Retriable errors - mark for wave-level retry, stop per-resource retry
			if isRetriableError(errStr) {
				return struct{}{}, backoff.Permanent(err)
			}

			// Other errors - retry with backoff
			return struct{}{}, err
		}
		return struct{}{}, nil
	}

	_, err := backoff.Retry(ctx, operation, backoff.WithBackOff(expBackoff), backoff.WithMaxTries(3))
	if err != nil {
		// Use lastErr if available for more accurate error message
		errToCheck := err
		if lastErr != nil {
			errToCheck = lastErr
		}
		errStr := errToCheck.Error()

		// Determine final state based on error type
		if isRetriableError(errStr) {
			r.SetState(PendingRetry)
		} else {
			r.SetState(Failed)
		}
		return errToCheck
	}

	r.SetState(Deleted)
	return nil
}

func (r Resources) NumOf(state ResourceState) int {
	count := 0
	for _, resource := range r {
		if resource.State() == state {
			count++
		}
	}
	return count
}

func (r Resources) VisibleCount() int {
	return len(r) - r.NumOf(Hidden)
}
