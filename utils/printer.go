package utils

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"

	"github.com/arafato/ali-nuke/types"
)

var (
	colorRed    = color.New(color.FgRed).SprintFunc()
	colorGreen  = color.New(color.FgGreen).SprintFunc()
	colorBlue   = color.New(color.FgBlue).SprintFunc()
	colorYellow = color.New(color.FgYellow).SprintFunc()
)

func PrintStatusWithContext(wg *sync.WaitGroup, ctx context.Context, resources types.Resources) {
	defer wg.Done()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			PrettyPrintStatus(resources)
		case <-ctx.Done():
			PrettyPrintStatus(resources)
			return
		}
	}
}

// colorizeStatus returns a colored status string based on the resource state
func colorizeStatus(state types.ResourceState) string {
	switch state {
	case types.Deleted:
		return colorGreen("Removed")
	case types.Filtered:
		return colorBlue("Filtered")
	case types.Removing, types.PendingRetry:
		return colorYellow("In-Progress")
	case types.Failed:
		return colorRed("Failed")
	default:
		return state.String()
	}
}

func PrettyPrintStatus(resources types.Resources) {
	data := [][]string{{"Region", "Product", "ID/Name", "Status"}}
	for _, resource := range resources {
		if resource.State() == types.Hidden {
			continue
		}

		status := colorizeStatus(resource.State())
		data = append(data, []string{resource.Region, resource.ProductName, resource.ResourceName, status})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header(data[0])
	table.Bulk(data[1:])
	table.Render()

	visibleCount := resources.VisibleCount()
	// Count PendingRetry as "In-Progress" for display
	inProgress := resources.NumOf(types.Removing) + resources.NumOf(types.PendingRetry)
	fmt.Printf("\nStatus: %d resources in total. %s %d, %s %d, %s %d, %s %d\n",
		visibleCount,
		colorGreen("Removed"), resources.NumOf(types.Deleted),
		colorYellow("In-Progress"), inProgress,
		colorBlue("Filtered"), resources.NumOf(types.Filtered),
		colorRed("Failed"), resources.NumOf(types.Failed))
}
