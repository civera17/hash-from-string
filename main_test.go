package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestExecuteDuplicateInput(t *testing.T) {
	input := []string{"LAXVIE", "LAXVIE", "LAXVIE"}

	var mu sync.Mutex
	callCounts := make(map[string]int)
	hashFn := func(s string) (int, error) {
		time.Sleep(3 *  time.Second) // Simulate a time-consuming operation
		mu.Lock()
		callCounts[s]++
		mu.Unlock()
		return 101, nil
	}

	results := execute(context.Background(), 3, input, hashFn)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0] != 101 {
		t.Fatalf("unexpected result: got %d, want %d", results[0], 101)
	}

	mu.Lock()
	defer mu.Unlock()
	if callCounts["LAXVIE"] != 1 {
		t.Fatalf("expected hashFn to be called once for duplicate input, got %d", callCounts["LAXVIE"])
	}
}

func TestExecuteDistinctInput(t *testing.T) {
	input := []string{"LAXVIE", "LAXFRA", "BOSNIS", "LONVIE"}
	expected := map[string]int{
		"LAXVIE": 111,
		"LAXFRA": 222,
		"BOSNIS": 333,
		"LONVIE": 444,
	}

	var mu sync.Mutex
	callCounts := make(map[string]int)
	hashFn := func(s string) (int, error) {
		mu.Lock()
		callCounts[s]++
		mu.Unlock()
		return expected[s], nil
	}

	results := execute(context.Background(), 2, input, hashFn)

	if len(results) != len(input) {
		t.Fatalf("expected %d results, got %d", len(input), len(results))
	}

	resultSet := make(map[int]int)
	for _, v := range results {
		resultSet[v]++
	}

	for str, want := range expected {
		if resultSet[want] == 0 {
			t.Fatalf("expected result %d for input %q was not produced", want, str)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	for str, count := range callCounts {
		if count != 1 {
			t.Fatalf("expected hashFn to be called once for %q, got %d", str, count)
		}
	}
}
