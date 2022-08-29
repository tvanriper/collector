package collector_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/tvanriper/collector"
)

func StringArrayMatches(l []string, r []string) bool {
	if len(r) != len(l) {
		return false
	}
	for i := 0; i < len(l); i++ {
		if l[i] != r[i] {
			return false
		}
	}
	return true
}

func IsNull(l interface{}) bool {
	return l == nil
}

func IsSame(l interface{}, r interface{}) bool {
	if IsNull(l) {
		return IsNull(r)
	}
	if IsNull(r) {
		// We already know l is not null.
		return false
	}

	if lar, ok := l.([]string); ok {
		if rar, ok := r.([]string); ok {
			return StringArrayMatches(lar, rar)
		}
		return false
	}

	if _, ok := r.([]string); ok {
		// We know l is not a collection of strings, so these don't match.
		return false
	}

	// Test int
	if li, ok := l.(int); ok {
		if ri, ok := r.(int); ok {
			return li == ri
		}
		return false
	}
	if _, ok := r.(int); ok {
		return false
	}

	// Test float64
	if lf, ok := l.(float64); ok {
		if rf, ok := r.(float64); ok {
			return lf == rf
		}
		return false
	}

	if _, ok := r.(float64); ok {
		return false
	}

	// Test string
	if ls, ok := l.(string); ok {
		if rs, ok := r.(string); ok {
			return ls == rs
		}
		return false
	}

	if _, ok := r.(string); ok {
		return false
	}
	return false
}

func TestCollector(t *testing.T) {
	storage := make([]interface{}, 0)

	coll := collector.New(1)

	chan1 := make(chan interface{}, 1)
	chan2 := make(chan interface{}, 1)
	defer close(chan1)
	defer close(chan2)

	// Create two depots using those channels.
	d1 := coll.Depot(chan1)
	d2 := coll.Depot(chan2)
	// And let them start listening for data.
	d1.Start()
	d2.Start()

	// This is probably overkill for a single thread.
	starting := sync.WaitGroup{}
	ending := sync.WaitGroup{}

	// Start the Collector's thread.  Our collection will go into storage.
	starting.Add(1)
	ending.Add(1)
	go func() {
		starting.Done()
		for {
			if data, more := <-coll.Listen(); more {
				storage = append(storage, data)
			} else {
				break
			}
		}
		ending.Done()
	}()

	sending1 := []interface{}{
		"Hello",
		1,
		nil,
		[]string{"This", "is", "data"},
		(float64)(3.141592653598),
	}

	sending2 := []interface{}{
		"Goodbye",
		12,
		"moo",
		[]string{"Woah", "why?"},
		(float64)(6.284185207179),
	}

	starting.Wait()

	// We'll zipper these.  The calls to Sleep help provide order to the
	// threads reading the values so we know how the collection should lay out.
	for i := 0; i < 5; i++ {
		t.Logf("Sending %v on an1, %v on an2", sending1[i], sending2[i])
		chan1 <- sending1[i]
		time.Sleep(10 * time.Millisecond)
		chan2 <- sending2[i]
		time.Sleep(10 * time.Millisecond)
	}

	// Close the collector, ending its thread.
	coll.Close()

	// Wait for the collector thread to end.
	ending.Wait()

	// Test the results
	if len(storage) < 10 {
		t.Logf("items: %v", storage)
		t.Fatalf("expected 10 items but received %d", len(storage))
	}
	for i := 0; i < 5; i++ {
		sel1 := i * 2
		sel2 := sel1 + 1
		expl := sending1[i]
		gotl := storage[sel1]
		expr := sending2[i]
		gotr := storage[sel2]
		if !IsSame(expl, gotl) {
			t.Errorf("expected %v but found %v", expl, gotl)
		}
		if !IsSame(expr, gotr) {
			t.Errorf("expected %v but found %v", expr, gotr)
		}
	}

	err := coll.Send("testing")
	if err == nil {
		t.Errorf("expected an error when sending to closed collector")
	}
}

func Example() {
	// Create a Collector.
	coll := collector.New(1)
	defer coll.Close()

	// Create two channels to demonstrate the collector.
	chan1 := make(chan interface{}, 1)
	chan2 := make(chan interface{}, 1)
	defer close(chan1)
	defer close(chan2)

	// Create two depots using those channels.
	d1 := coll.Depot(chan1)
	d2 := coll.Depot(chan2)
	// And let them start listening for data.
	d1.Start()
	defer d1.Close()
	d2.Start()
	defer d2.Close()

	// Create a go thread of our collector listening to information from the
	// channels.
	go func() {
		for {
			if data, more := <-coll.Listen(); more {
				if s, ok := data.(string); ok {
					fmt.Printf("Incoming: %s\n", s)
				}
			} else {
				break
			}
		}
	}()

	// Wait for the thread to get going.
	time.Sleep(50 * time.Millisecond)

	// Adding a little time between requests to make the output predictable.
	chan1 <- "Hello"
	time.Sleep(10 * time.Millisecond)
	chan2 <- "World"
	time.Sleep(10 * time.Millisecond)
	chan1 <- "Goodbye!"
	time.Sleep(10 * time.Millisecond)

	// Output:
	// Incoming: Hello
	// Incoming: World
	// Incoming: Goodbye!
}
