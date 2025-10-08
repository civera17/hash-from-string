package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

func hash(s string) (int, error) {
	time.Sleep(3 * time.Second) // Simulate a time-consuming operation
	return rand.Int(), nil
}

func main() {
	input := []string{
		"LAXVIE",
		"LAXVIE",
		"LAXVIE",
		"LAXVIE",
		"LAXVIE",
		"LAXVIE",
	}
	const numWorkers = 5
	results := execute(context.Background(), numWorkers, input, hash)
	for _, r := range results {
		fmt.Printf("reader received result: %v\n", r)
	}
}

func execute(parent context.Context, numWorkers int, input []string, hashFn func(string) (int, error)) []int {
	jobs := make(chan string)
	results := make(chan int)

	var sm sync.Map
	var jobWG sync.WaitGroup
	var workerWG sync.WaitGroup
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	startTheWork(ctx, numWorkers, jobs, results, &jobWG, &sm, &workerWG, hashFn)

	jobWG.Add(len(input))
	// Here was the issue feed with jobs in a separate goroutine
	go func() {
		for _, v := range input {
			jobs <- v
		}
		close(jobs)
	}()

	// And wait for all jobs to be processed in a separate goroutine
	go func() {
		jobWG.Wait()
		close(results)
		cancel()
	}()

	var output []int
	for v := range results {
		output = append(output, v)
	}
	workerWG.Wait()
	return output
}

func startTheWork(ctx context.Context, workers int, j <-chan string, out chan<- int, wg *sync.WaitGroup, sm *sync.Map, workerWG *sync.WaitGroup, hashFn func(string) (int, error)) {
	for i := range workers {
		workerWG.Add(1)
		go func() {
			defer workerWG.Done()
			for {
				select {
				case <-ctx.Done():
					fmt.Printf("worker %d is stopping\n", i)
					return
				case v, ok := <-j:
					if !ok {
						fmt.Printf("worker %d is stopping; jobs channel closed\n", i)
						return
					}
					fmt.Printf("worker %d received job: %v\n", i, v)
					// Check if result is already cached
					if _, exists := sm.Load(v); exists {
						fmt.Printf("worker %d found cached result for %v, skipping\n", i, v)
						wg.Done()
						continue
					}
					// Not cached, process the job, avoid other goroutines doing the same work
					sm.Store(v, struct{}{}) // Placeholder to indicate processing
					fmt.Printf("worker %d is processing job: %v\n", i, v)
					n, err := hashFn(v)
					if err != nil {
						fmt.Printf("worker %d encountered error: %v\n", i, err)
						wg.Done()
						continue
					}
					sm.Store(v, n)
					out <- n
					wg.Done()
				}
			}
		}()
	}
}
