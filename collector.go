package collector

import (
	"fmt"
	"sync"
)

// Collector provides a structure to which all Depots send their data.
type Collector struct {
	buffer  int
	emitter chan interface{}
	closed  bool
	closers map[*Depot]chan struct{}
	lock    sync.Mutex
}

// Depot provides a structure through which channels of data may send their
// data to a Collector.
type Depot struct {
	collector *Collector
	ch        <-chan interface{}
	closer    chan struct{}
}

// New creates a new Collector.
func New(size int) *Collector {
	return &Collector{
		buffer:  size,
		emitter: make(chan interface{}, size),
		closed:  false,
		closers: make(map[*Depot]chan struct{}),
	}
}

// Listen provides a way for something to listen to messages from the
// Collector's channel.
func (c *Collector) Listen() <-chan interface{} {
	return c.emitter
}

// Send provides a way to send data through the Collector's channel.
func (c *Collector) Send(data interface{}) (err error) {
	if !c.closed {
		c.emitter <- data
		return nil
	}
	return fmt.Errorf("collector is closed")
}

// Close ends the Collector, preventing it from sending more data.
func (c *Collector) Close() {
	c.closed = true
	close(c.emitter)
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, closer := range c.closers {
		close(closer)
	}
	c.closers = make(map[*Depot]chan struct{}, 0)
}

// Depot creates a new Depot receiving its data from the provided channel.  All
// data received through the provided channel is resent through the Collector.
func (c *Collector) Depot(channel <-chan interface{}) *Depot {
	closer := make(chan struct{})
	result := &Depot{
		collector: c,
		ch:        channel,
		closer:    closer,
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.closers[result] = closer
	return result
}

// Start executes Depot's Run function in its own Go thread.
func (d *Depot) Start() {
	go func() {
		d.Run()
	}()
}

// Run waits for data from the channel and sends it to the collector.  It ends
// automatically when something closes the channel.  It will also remove itself
// from the collector.
func (d *Depot) Run() {
	func() {
		for {
			select {
			case data, ok := <-d.ch:
				if ok {
					d.collector.Send(data)
				} else {
					return
				}
			case <-d.closer:
				return
			}
		}
	}()
	d.remove()
}

func (d *Depot) remove() {
	if d.collector != nil {
		d.collector.lock.Lock()
		defer d.collector.lock.Unlock()
		delete(d.collector.closers, d)
		d.collector = nil
	}
}

// Close ensures this Depot tranfers no more data to its Collector.
func (d *Depot) Close() {
	if d.collector != nil {
		d.remove()
		close(d.closer)
	}
}

// Send writes data to this Depot's Collector for transmission, without going
// through the channel.
func (d *Depot) Send(data interface{}) error {
	if d.collector == nil {
		return fmt.Errorf("unable to send on closed Depot")
	}
	return d.collector.Send(data)
}
