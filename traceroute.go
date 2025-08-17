// Copyright 2025 icmpkg Author. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//      http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package icmpkg

import (
	"context"
	"fmt"
	logpkg "log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Global variables for ICMP ID generation and debug/trace logging.
var (
	icmpId          = uint32(os.Getpid() & 0xffff)         // Initial ICMP ID derived from process ID, masked to 16 bits.
	tracerouteDebug = os.Getenv("TRACEROUTE_DEBUG") == "T" // Enables debug logging if TRACEROUTE_DEBUG is set to "T".
	tracerouteTrace = os.Getenv("TRACEROUTE_TRACE") == "T" // Enables trace logging if TRACEROUTE_TRACE is set to "T".
)

// nextIcmpId generates the next ICMP ID, incrementing atomically and wrapping around at 2^15.
func nextIcmpId() uint32 {
	return atomic.AddUint32(&icmpId, 1) % (2 << 15)
}

// traceroute manages ICMP-based ping or traceroute operations with configuration and synchronization.
type traceroute struct {
	lo                    *logpkg.Logger    // Logger instance for debug and trace output.
	address               string            // Target address for ping/traceroute.
	addr                  net.Addr          // Resolved network address of the target.
	ip4                   string            // IPv4 address as a string.
	maxTTL, maxHop, count int               // Maximum TTL, maximum hops, and number of packets to send.
	writeDur, readDur     time.Duration     // Durations for write and read timeouts.
	wc, rc, hc            chan *Proto       // Channels for writing, reading, and handling Proto messages.
	id                    []int             // Array of ICMP IDs for each TTL.
	ic                    []chan *Proto     // Array of channels for receiving Proto messages per TTL.
	pec, hec, cec         chan struct{}     // Channels for signaling pong, handler, and context termination.
	runOnce, stopOnce     *sync.Once        // Ensure Run and Stop are executed only once.
	exit                  bool              // Flag to indicate termination.
	pongHandler           func(pong *Proto) // Optional callback for handling pong responses.
	ctx                   context.Context   // Context for cancellation.
	packet                *packet           // Packet handler for ICMP communication.
	wg                    *sync.WaitGroup   // WaitGroup for synchronizing goroutines.
	traceroute            bool              // Flag to indicate traceroute (true) or ping (false) mode.
}

// Traceroute creates a traceroute instance with default write and read durations of 500ms.
func Traceroute(address string, maxTTL, count int) *traceroute {
	// Initialize traceroute with default durations for write and read operations.
	return TracerouteDuration(address, maxTTL, count, time.Millisecond*500, time.Millisecond*500)
}

// TracerouteDuration creates a traceroute instance with specified write and read durations.
func TracerouteDuration(address string, maxTTL, count int, writeDur, readDur time.Duration) *traceroute {
	// Initialize a new traceroute instance with the provided parameters and traceroute mode enabled.
	return newTraceroute(address, maxTTL, count, writeDur, readDur, true)
}

// newTraceroute initializes a traceroute instance with the given configuration.
func newTraceroute(address string, maxTTL, count int, writeDur, readDur time.Duration, route bool) *traceroute {
	tr := &traceroute{
		address:    address,                     // Set target address.
		maxTTL:     maxTTL,                      // Set maximum TTL.
		maxHop:     maxTTL,                      // Set maximum hops (initially equal to maxTTL).
		count:      count,                       // Set number of packets to send per TTL.
		writeDur:   writeDur,                    // Set write timeout duration.
		readDur:    readDur,                     // Set read timeout duration.
		wc:         make(chan *Proto, 1),        // Initialize write channel.
		rc:         make(chan *Proto, 1),        // Initialize read channel.
		hc:         make(chan *Proto, 1),        // Initialize handler channel.
		id:         make([]int, maxTTL),         // Initialize ICMP ID array.
		ic:         make([]chan *Proto, maxTTL), // Initialize per-TTL Proto channels.
		pec:        make(chan struct{}, 1),      // Initialize pong exit channel.
		hec:        make(chan struct{}, 1),      // Initialize handler exit channel.
		runOnce:    &sync.Once{},                // Initialize Run once guard.
		stopOnce:   &sync.Once{},                // Initialize Stop once guard.
		wg:         &sync.WaitGroup{},           // Initialize WaitGroup for goroutine synchronization.
		traceroute: route,                       // Set traceroute or ping mode.
	}
	// Resolve the target address and its IPv4 string representation.
	tr.addr, tr.ip4 = ip4(address)
	// Set up logger for ping mode if debug or trace is enabled.
	if !route && (pingDebug || pingTrace) {
		tr.lo = logpkg.New(os.Stdout, fmt.Sprintf("[ping:%-24s] ", tr.address), logpkg.LstdFlags)
	}
	// Set up logger for traceroute mode if debug or trace is enabled.
	if route && (tracerouteDebug || tracerouteTrace) {
		tr.lo = logpkg.New(os.Stdout, fmt.Sprintf("[route:%-23s] ", tr.address), logpkg.LstdFlags)
	}
	return tr
}

// debug logs a debug message if debug mode is enabled for ping or traceroute.
func (tr *traceroute) debug(format string, arg ...any) {
	if tr.traceroute && tracerouteDebug {
		tr.lo.Println(fmt.Sprintf(format, arg...)) // Log debug message in traceroute mode.
	}
	if !tr.traceroute && pingDebug {
		tr.lo.Println(fmt.Sprintf(format, arg...)) // Log debug message in ping mode.
	}
}

// trace logs a trace message if trace mode is enabled for ping or traceroute.
func (tr *traceroute) trace(format string, arg ...any) {
	if tr.traceroute && tracerouteTrace {
		tr.lo.Println(fmt.Sprintf(format, arg...)) // Log trace message in traceroute mode.
	}
	if !tr.traceroute && pingTrace {
		tr.lo.Println(fmt.Sprintf(format, arg...)) // Log trace message in ping mode.
	}
}

// Addr returns the resolved network address of the target.
func (tr *traceroute) Addr() net.Addr { return tr.addr }

// Ip4 returns the IPv4 address of the target as a string.
func (tr *traceroute) Ip4() string { return tr.ip4 }

// Context sets the context for cancellation and initializes the context exit channel.
func (tr *traceroute) Context(ctx context.Context) {
	tr.ctx = ctx
	tr.cec = make(chan struct{}, 1)
}

// PongHandler sets the callback function for handling pong responses.
func (tr *traceroute) PongHandler(handler func(pong *Proto)) {
	tr.pongHandler = handler
}

// Run starts the traceroute or ping operation, ensuring it runs only once.
func (tr *traceroute) Run() {
	fn := func() {
		tr.trace("Run() start")             // Log start of Run operation.
		defer tr.trace("Run() end")         // Log end of Run operation.
		tr.packet = newPacket(tr.rc, tr.wc) // Initialize packet handler.
		go tr.startPong()                   // Start pong processing goroutine.
		go tr.startHandler()                // Start handler goroutine.
		go tr.startCtx()                    // Start context monitoring goroutine.
		tr.runPing()                        // Run the ping or traceroute operation.
		tr.Stop()                           // Stop the operation after completion.
	}
	tr.runOnce.Do(fn) // Ensure Run is executed only once.
}

// Stop terminates the traceroute or ping operation, ensuring it stops only once.
func (tr *traceroute) Stop() {
	fn := func() {
		tr.trace("Stop() start")      // Log start of Stop operation.
		defer tr.trace("Stop() end")  // Log end of Stop operation.
		tr.exit = true                // Set exit flag.
		tr.packet.stop()              // Stop the packet handler.
		tr.pec <- struct{}{}          // Signal pong goroutine to exit.
		close(tr.pec)                 // Close pong exit channel.
		tr.trace("Stop() closed pec") // Log pong channel closure.
		tr.hec <- struct{}{}          // Signal handler goroutine to exit.
		close(tr.hec)                 // Close handler exit channel.
		tr.trace("Stop() closed hec") // Log handler channel closure.
		if tr.cec != nil {
			tr.cec <- struct{}{}          // Signal context goroutine to exit.
			close(tr.cec)                 // Close context exit channel.
			tr.trace("Stop() closed cec") // Log context channel closure.
		}
		tr.closes() // Close all per-TTL channels.
	}
	tr.stopOnce.Do(fn) // Ensure Stop is executed only once.
}

// pong processes a received Proto message and forwards it to the appropriate TTL channel.
func (tr *traceroute) pong(pto *Proto) {
	tr.trace("pong() start")     // Log start of pong processing.
	defer tr.trace("pong() end") // Log end of pong processing.
	ttl := pto.TTL
	if tr.traceroute {
		ttl-- // Adjust TTL index for traceroute mode.
	}
	tr.ic[ttl] <- pto // Send Proto to the corresponding TTL channel.
}

// startPong runs a goroutine to process incoming Proto messages from the read channel.
func (tr *traceroute) startPong() {
	tr.trace("startPong() start")     // Log start of pong goroutine.
	defer tr.trace("startPong() end") // Log end of pong goroutine.
	for {
		select {
		case <-tr.pec:
			return // Exit if pong exit channel is signaled.
		case pto, ok := <-tr.rc:
			if !ok {
				return // Exit if read channel is closed.
			}
			tr.debug("packet->>>>>>: %s", pto.String()) // Log received Proto message.
			if tr.traceroute && pto.Ip4 == tr.ip4 && tr.maxHop > pto.TTL {
				tr.trace("found max hop: %d", pto.TTL) // Update max hop if destination reached.
				tr.maxHop = pto.TTL
			}
			tr.pong(pto) // Process the Proto message.
		}
	}
}

// handler forwards a Proto message to the handler channel and invokes the pong handler.
func (tr *traceroute) handler(pto *Proto) {
	if tr.exit {
		return // Skip if operation is terminated.
	}
	tr.hc <- pto                       // Send Proto to handler channel.
	tr.debug("handler<<<<<-: %s", pto) // Log handled Proto message.
}

// startHandler runs a goroutine to process Proto messages from the handler channel.
func (tr *traceroute) startHandler() {
	tr.trace("startHandler() start")     // Log start of handler goroutine.
	defer tr.trace("startHandler() end") // Log end of handler goroutine.
	for {
		select {
		case <-tr.hec:
			return // Exit if handler exit channel is signaled.
		case pto, ok := <-tr.hc:
			if !ok {
				return // Exit if handler channel is closed.
			}
			if tr.pongHandler != nil && pto != nil {
				tr.pongHandler(pto) // Invoke pong handler callback if set.
			}
		}
	}
}

// closes closes all per-TTL Proto channels.
func (tr *traceroute) closes() {
	for ttl, ic := range tr.ic {
		if ic != nil {
			close(ic) // Close the TTL-specific channel.
			if tr.traceroute {
				tr.trace("closes() closed ic ttl: %d", ttl+1) // Log closure in traceroute mode.
			} else {
				tr.trace("closes() closed ic") // Log closure in ping mode.
			}
		}
	}
}

// ping sends a Proto message to the write channel for transmission.
func (tr *traceroute) ping(pto *Proto) {
	if tr.exit {
		return // Skip if operation is terminated.
	}
	tr.wc <- pto                       // Send Proto to write channel.
	tr.debug("packet<<<<<<-: %s", pto) // Log sent Proto message.
}

// runPing executes the ping or traceroute operation for each TTL.
func (tr *traceroute) runPing() {
	tr.trace("runPing() start")     // Log start of runPing operation.
	defer tr.trace("runPing() end") // Log end of runPing operation.

	closes := func() {
		close(tr.wc)                    // Close write channel.
		tr.trace("runPing() closed wc") // Log write channel closure.
		close(tr.hc)                    // Close handler channel.
		tr.trace("runPing() closed hc") // Log handler channel closure.
	}

	for ttl := 0; ttl < tr.maxHop; ttl++ {
		if tr.id[ttl] == 0 {
			tr.id[ttl] = int(nextIcmpId())    // Assign a new ICMP ID for the TTL.
			tr.ic[ttl] = make(chan *Proto, 1) // Initialize Proto channel for the TTL.
		}
		id := tr.id[ttl]
		ttl0 := ttl
		if tr.traceroute {
			ttl0++ // Adjust TTL for traceroute mode.
		}
		if tr.exit {
			closes() // Close channels if operation is terminated.
			return
		}
		tr.ping(pingProto(ttl0, id, 0, tr.addr, tr.ip4)) // Send initial ping for the TTL.
		tr.handler(tr.readTTL(ttl, id, 0))               // Process response for initial ping.
		tr.wg.Add(1)                                     // Increment WaitGroup for TTL goroutine.
		go tr.runTTL(ttl, tr.count)                      // Start goroutine for remaining pings in TTL.
		if !tr.traceroute {
			break // Exit loop after first TTL in ping mode.
		}
	}
	tr.wg.Wait() // Wait for all TTL goroutines to complete.
	closes()     // Close channels after completion.
}

// runTTL sends additional pings for a specific TTL and processes responses.
func (tr *traceroute) runTTL(ttl, count int) {
	ttl0 := ttl
	if tr.traceroute {
		ttl0++ // Adjust TTL for traceroute mode.
	}
	tr.trace("runTTL() start ttl: %d count: %d", ttl0, count)     // Log start of runTTL.
	defer tr.trace("runTTL() end ttl: %d count: %d", ttl0, count) // Log end of runTTL.
	defer tr.wg.Done()                                            // Signal WaitGroup completion.
	for seq := 1; seq < count; seq++ {
		if tr.exit {
			return // Exit if operation is terminated.
		}
		tr.ping(pingProto(ttl0, tr.id[ttl], seq, tr.addr, tr.ip4)) // Send ping for sequence.
		tr.handler(tr.readTTL(ttl, tr.id[ttl], seq))               // Process response.
	}
}

// readTTL waits for a response for a specific TTL, ID, and sequence number, handling timeouts.
func (tr *traceroute) readTTL(ttl, id, seq int) (pto *Proto) {
	now := time.Now()
	ttl0 := ttl
	if tr.traceroute {
		ttl0++ // Adjust TTL for traceroute mode.
	}
	tr.trace("readTTL() start ttl: %d id: %d seq: %d", ttl0, id, seq)     // Log start of readTTL.
	defer tr.trace("readTTL() end ttl: %d id: %d seq: %d", ttl0, id, seq) // Log end of readTTL.
	for {
		select {
		case pto = <-tr.ic[ttl]:
			if seq > 0 {
				time.Sleep(tr.readDur - time.Since(now)) // Adjust timing for subsequent pings.
			}
			return // Return received Proto message.
		case <-time.After(tr.readDur):
			pto = timeoutProto(ttl0, id, seq)                                   // Create timeout Proto on read timeout.
			tr.trace("readTTL() timeout ttl: %d id: %d seq: %d", ttl0, id, seq) // Log timeout.
			tr.debug("timeout->>>>>: %s", pto)                                  // Log timeout Proto.
			return                                                              // Return timeout Proto.
		}
	}
}

// startCtx runs a goroutine to monitor the context for cancellation.
func (tr *traceroute) startCtx() {
	if tr.ctx == nil {
		return // Skip if no context is set.
	}
	tr.trace("startCtx() start")     // Log start of context monitoring.
	defer tr.trace("startCtx() end") // Log end of context monitoring.
	go func() {
		for {
			select {
			case <-tr.cec:
				return // Exit if context exit channel is signaled.
			case <-tr.ctx.Done():
				tr.Stop() // Stop operation on context cancellation.
				return
			}
		}
	}()
}

// ip4 resolves an address to an IPv4 net.Addr and its string representation.
func ip4(s string) (net.Addr, string) {
	if ip := net.ParseIP(s); ip != nil {
		return &net.IPAddr{IP: ip}, s // Return parsed IP address if valid.
	}
	addr, _ := net.ResolveIPAddr("ip4", s) // Resolve address to IPv4.
	return addr, aip4(addr)                // Return resolved address and its string form.
}

// aip4 converts a net.Addr to its IPv4 string representation.
func aip4(a net.Addr) (ip4 string) {
	if a == nil {
		return // Return empty string if address is nil.
	}
	ipa, ok := a.(*net.IPAddr)
	if !ok || ipa == nil {
		return // Return empty string if not an IPAddr or nil.
	}
	return ipa.String() // Return IPv4 address as string.
}
