# icmpkg

`icmpkg` is a Go package for performing ICMP-based ping and traceroute operations. 

It provides a flexible and thread-safe implementation for sending and receiving ICMP Echo Request and Reply messages, with support for configurable TTL (Time To Live), packet counts, and timeouts. 

The package supports both ping and traceroute modes, making it suitable for network diagnostics and analysis.

## Features

- **Ping and Traceroute Support**: Perform standard ping operations or trace the route to a destination with configurable TTL and packet counts.
- **Customizable Timeouts**: Set write and read durations for fine-grained control over operation timing.
- **Thread-Safe Design**: Uses mutexes and atomic operations to ensure safe concurrent packet handling.
- **Context Support**: Allows cancellation of operations using Go's context package.
- **Custom Pong Handlers**: Define custom callbacks to process ICMP responses.
- **Debug and Trace Logging**: Enable detailed logging using environment variables for debugging and tracing.

## Installation

To use `icmpkg` in your Go project, ensure you have Go installed and run:

```bash
go get github.com/go-the-way/icmpkg
```

## Usage

### Ping Example

Perform a ping operation to a target address with 3 packets and default timeouts (500ms):

```go
package main

import (
	"fmt"
	
	"github.com/go-the-way/icmpkg"
)

func main() {
	// Create a ping instance targeting 8.8.8.8 with 3 packets.
	ping := icmpkg.Ping("8.8.8.8", 3)

	// Set a custom handler to process pong responses.
	ping.PongHandler(func(pong *icmpkg.Proto) {
		fmt.Printf("Ping response: %s\n", pong.String())
	})

	// Run the ping operation.
	ping.Run()
}
```

### Traceroute Example

Perform a traceroute operation to a target address with a maximum TTL of 30 and 3 packets per TTL:

```go
package main

import (
	"fmt"
	
	"github.com/go-the-way/icmpkg"
)

func main() {
	// Create a traceroute instance targeting 8.8.8.8 with max TTL 30 and 3 packets per TTL.
	tr := icmpkg.Traceroute("8.8.8.8", 30, 3)

	// Set a custom handler to process pong responses.
	tr.PongHandler(func(pong *icmpkg.Proto) {
		fmt.Printf("Traceroute hop: %s\n", pong.String())
	})

	// Run the traceroute operation.
	tr.Run()
}
```

### Context Cancellation

Use a context to cancel the operation after a timeout:

```go
package main

import (
	"context"
	"fmt"
	"time"
	
	"github.com/go-the-way/icmpkg"
)

func main() {
	// Create a context with a 5-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Create a ping instance.
	ping := icmpkg.Ping("8.8.8.8", 10)
	
	// Set the context for cancellation.
	ping.Context(ctx)
	
	// Set a pong handler.
	ping.PongHandler(func(pong *icmpkg.Proto) {
		fmt.Printf("Ping response: %s\n", pong.String())
	})
	
	// Run the ping operation.
	ping.Run()
}
```

## Environment Variables

The package supports debug and trace logging controlled by environment variables:

- `ICMPKG_DEBUG=T`: Enable debug logging for low-level packet operations.
- `ICMPKG_TRACE=T`: Enable trace logging for low-level packet operations.
- `PING_DEBUG=T`: Enable debug logging for ping operations.
- `PING_TRACE=T`: Enable trace logging for ping operations.
- `TRACEROUTE_DEBUG=T`: Enable debug logging for traceroute operations.
- `TRACEROUTE_TRACE=T`: Enable trace logging for traceroute operations.

Example to enable debug logging:

```bash
export PING_DEBUG=T
go run your_program.go
```

## Package Structure

- `Proto`: Struct representing an ICMP packet's metadata, including TTL, ID, sequence number, address, and RTT.
- `packet`: Internal struct for managing low-level ICMP packet sending and receiving.
- `traceroute`: Core struct for ping and traceroute operations, handling TTL iteration, packet sending, and response processing.
- `Ping` and `Traceroute`: High-level functions to initialize ping or traceroute operations.
- `PingDuration` and `TracerouteDuration`: Variants allowing custom write and read timeouts.

## Requirements

- Go 1.18 or later.
- Root/administrator privileges may be required for raw ICMP socket operations on some systems.
- IPv4 network support (the package uses `ip4:icmp` protocol).

## Notes

- The package uses `golang.org/x/net/icmp` and `golang.org/x/net/ipv4` for low-level ICMP communication.
- Ensure proper error handling in production code, as the provided examples omit some error checks for brevity.
- The package is designed for IPv4; IPv6 support is not currently implemented.

## License

This package is licensed under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).