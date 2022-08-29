# Collector

[![Go Reference](https://pkg.go.dev/badge/golang.org/x/example.svg)](https://pkg.go.dev/github.com/tvanriper/collector)
[![Coverage Status](https://coveralls.io/repos/github/tvanriper/collector/badge.svg?branch=main)](https://coveralls.io/github/tvanriper/collector?branch=main)

This library provides a way to collect different channels of information to a
single channel.  The library is type-agnostic, using the interface{} type for
data transmission, to provide as much flexibility as possible.

## Usage

Generally, you would do the following:

* Create a new Collector.
* Create a depot for each channel with information you want to receive.
* Start each depot or Run it in your own thread, as desired, to start receiving
  information.
* Close the collector when finished.  This also ends all Depot threads.
* Also Close a Depot when you no longer want it to receive information, or
  when you're through.

## Notes

Depots attempting to send to a closed Collector will not panic, but the
Collector will not receive the data.

Channels given to a depot that someone closes automatically ends the Run or
Start thread for that Depot.

If you do not care to use a depot, you could use the Collector's Send function
instead.  You'd gain a benefit of tracking if the channel is closed, gracefully
failing in such cases rather than experiencing a panic.

You can use github.com/tvanriper/announcer to do the opposite of what this
library does, in that it can take a single channel and send it to many
listeners.  The two libraries are meant to complement each other.

## Installation

`go get github.com/tvanriper/collector`
