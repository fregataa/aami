package ssh

import (
	"context"
	"sync"
)

// RunParallel executes a command on multiple nodes in parallel
func (e *Executor) RunParallel(ctx context.Context, nodes []Node, command string) []Result {
	results := make([]Result, len(nodes))
	sem := make(chan struct{}, e.config.MaxParallel)
	var wg sync.WaitGroup

	for i, node := range nodes {
		wg.Add(1)
		go func(idx int, n Node) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[idx] = Result{Node: n.Name, Error: ctx.Err()}
				return
			}

			results[idx] = e.RunWithRetry(ctx, n, command)
		}(i, node)
	}

	wg.Wait()
	return results
}

// RunParallelWithCallback executes a command on multiple nodes and calls a callback for each result
func (e *Executor) RunParallelWithCallback(ctx context.Context, nodes []Node, command string, callback func(Result)) {
	sem := make(chan struct{}, e.config.MaxParallel)
	var wg sync.WaitGroup

	for _, node := range nodes {
		wg.Add(1)
		go func(n Node) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				callback(Result{Node: n.Name, Error: ctx.Err()})
				return
			}

			result := e.RunWithRetry(ctx, n, command)
			callback(result)
		}(node)
	}

	wg.Wait()
}

// Summary provides a summary of parallel execution results
type Summary struct {
	Total     int
	Succeeded int
	Failed    int
	Results   []Result
}

// RunParallelWithSummary executes a command on multiple nodes and returns a summary
func (e *Executor) RunParallelWithSummary(ctx context.Context, nodes []Node, command string) Summary {
	results := e.RunParallel(ctx, nodes, command)

	summary := Summary{
		Total:   len(results),
		Results: results,
	}

	for _, r := range results {
		if r.Error == nil {
			summary.Succeeded++
		} else {
			summary.Failed++
		}
	}

	return summary
}
