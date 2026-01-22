package types

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"

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
)

// State returns the current state of the resource (thread-safe)
func (r *Resource) State() ResourceState {
	return ResourceState(r.state.Load())
}

// SetState sets the state of the resource (thread-safe)
func (r *Resource) SetState(s ResourceState) {
	r.state.Store(int32(s))
}

func (r *Resource) Remove(ctx context.Context) error {
	operation := func() (struct{}, error) {
		r.SetState(Removing)
		err := r.Removable.Remove(r.Region, r.ResourceID, r.ResourceName)
		if err != nil {
			// Check for permanent errors that shouldn't be retried
			errStr := err.Error()
			if strings.Contains(errStr, "403") ||
				strings.Contains(errStr, "Forbidden") ||
				strings.Contains(errStr, "InvalidAccessKeyId") {
				return struct{}{}, backoff.Permanent(errors.New("unauthorized request: " + errStr))
			}
			return struct{}{}, err
		}
		return struct{}{}, nil
	}

	_, err := backoff.Retry(ctx, operation, backoff.WithBackOff(backoff.NewExponentialBackOff()), backoff.WithMaxTries(3))
	if err != nil {
		r.SetState(Failed)
		return err
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
