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
	"fmt"
	logpkg "log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Constants defining the network protocol and listening address for ICMP communication.
const (
	listenNetwork = "ip4:icmp" // Specifies the ICMP over IPv4 network protocol.
	listenAddress = "0.0.0.0"  // Listening address to accept all incoming connections.
)

// Global variables controlling debug and trace logging based on environment variables.
var (
	icmpkgDebug = func() bool { return os.Getenv("ICMPKG_DEBUG") == "T" } // Enables debug logging if ICMPKG_DEBUG is set to "T".
	icmpkgTrace = func() bool { return os.Getenv("ICMPKG_TRACE") == "T" } // Enables trace logging if ICMPKG_TRACE is set to "T".
)

// ttlOpt stores TTL (Time To Live) and timestamp information for a packet.
type ttlOpt struct {
	ttl  int   // Time To Live value for the packet.
	unix int64 // Unix timestamp in milliseconds when the packet was sent.
}

// packet represents an ICMP packet handler with connection, logging, and synchronization primitives.
type packet struct {
	lo         *logpkg.Logger    // Logger instance for debug and trace output.
	packetConn *icmp.PacketConn  // ICMP packet connection for sending and receiving packets.
	wc         chan<- *Proto     // Write channel for sending Proto messages.
	rc         <-chan *Proto     // Read channel for receiving Proto messages.
	mu         *sync.Mutex       // Mutex for thread-safe access to the TTL map.
	m          map[string]ttlOpt // Map storing TTL and timestamp for packets, keyed by ID-Seq.
	wec, rec   chan struct{}     // Channels for signaling write and read goroutine termination.
}

// newPacket creates and initializes a new packet handler instance.
func newPacket(wc chan<- *Proto, rc <-chan *Proto) *packet {
	pkt := &packet{
		wc:  wc,                      // Initialize write channel.
		rc:  rc,                      // Initialize read channel.
		mu:  &sync.Mutex{},           // Initialize mutex for thread safety.
		m:   make(map[string]ttlOpt), // Initialize TTL map.
		wec: make(chan struct{}, 1),  // Initialize write exit channel with buffer size 1.
		rec: make(chan struct{}, 1),  // Initialize read exit channel with buffer size 1.
	}
	// Set up logger if debug or trace mode is enabled.
	if icmpkgDebug() || icmpkgTrace() {
		pkt.lo = logpkg.New(os.Stdout, fmt.Sprintf("[icmp-packet%0-18s] ", ""), logpkg.LstdFlags)
	}
	// Start the packet handler's main loop.
	pkt.run()
	return pkt
}

// debug logs a debug message if debug mode is enabled.
func (p *packet) debug(format string, arg ...any) {
	if icmpkgDebug() {
		p.lo.Println(fmt.Sprintf(format, arg...)) // Log formatted debug message.
	}
}

// trace logs a trace message if trace mode is enabled.
func (p *packet) trace(format string, arg ...any) {
	if icmpkgTrace() {
		p.lo.Println(fmt.Sprintf(format, arg...)) // Log formatted trace message.
	}
}

// listen sets up the ICMP packet connection to listen on the specified network and address.
func (p *packet) listen() {
	p.trace("listen() start")     // Log start of listen operation.
	defer p.trace("listen() end") // Log end of listen operation.
	var err error
	// Create an ICMP packet connection.
	p.packetConn, err = icmp.ListenPacket(listenNetwork, listenAddress)
	if err != nil {
		// Panic if listening fails, including error details.
		panic(fmt.Sprintf("listen() listen on[%s:%s] error:%v", listenNetwork, listenAddress, err))
		return
	}
	// Log successful listening setup.
	p.trace("listen() listen on %s:%s", listenNetwork, listenAddress)
}

// run initializes the packet handler by setting up the listener and starting read/write goroutines.
func (p *packet) run() {
	p.trace("run() start")     // Log start of run operation.
	defer p.trace("run() end") // Log end of run operation.
	p.listen()                 // Set up ICMP listener.
	p.start()                  // Start read and write goroutines.
}

// start launches separate goroutines for reading and writing ICMP packets.
func (p *packet) start() {
	p.trace("start() start")     // Log start of start operation.
	defer p.trace("start() end") // Log end of start operation.
	go p.startWrite()            // Start write goroutine.
	go p.startRead()             // Start read goroutine.
}

// stop terminates the read and write goroutines and closes the packet connection.
func (p *packet) stop() {
	p.trace("stop() start")     // Log start of stop operation.
	defer p.trace("stop() end") // Log end of stop operation.
	p.wec <- struct{}{}         // Signal write goroutine to exit.
	close(p.wec)                // Close write exit channel.
	p.rec <- struct{}{}         // Signal read goroutine to exit.
	close(p.rec)                // Close read exit channel.
	if p.packetConn != nil {
		_ = p.packetConn.Close() // Close the ICMP packet connection.
	}
}

// startWrite handles writing ICMP packets to the network.
func (p *packet) startWrite() {
	p.trace("startWrite() start")     // Log start of write operation.
	defer p.trace("startWrite() end") // Log end of write operation.
	for {
		select {
		case <-p.wec:
			return // Exit if write exit channel is signaled.
		case pto, ok := <-p.rc:
			if !ok {
				return // Exit if read channel is closed.
			}
			setTtl := pto.TTL > 0 // Check if TTL needs to be set.
			if setTtl {
				// Set TTL for the packet connection.
				if err := p.packetConn.IPv4PacketConn().SetTTL(pto.TTL); p.closed(err) {
					return // Exit if connection is closed.
				}
			}
			// Write packet data to the destination address.
			_, err := p.packetConn.WriteTo(pto.buf(), pto.Addr)
			if err != nil {
				// Log error if write fails.
				p.debug("conn<<<<<<-err: %s, %v", pto, err)
				if p.closed(err) {
					return // Exit if connection is closed.
				}
			} else {
				// Log successful write and store TTL information.
				p.debug("conn<<<<<<-ok: %s", pto)
				p.setTTL(pto.TTL, pto.ID, pto.Seq)
			}
		}
	}
}

// startRead handles reading ICMP packets from the network.
func (p *packet) startRead() {
	p.trace("startRead() start")     // Log start of read operation.
	defer p.trace("startRead() end") // Log end of read operation.
	buf := make([]byte, 64)          // Buffer for reading ICMP packets.
	for {
		select {
		case <-p.rec:
			close(p.wc)                      // Close write channel on exit.
			p.trace("startRead() closed wc") // Log write channel closure.
			return
		default:
			// Set a read deadline to prevent blocking indefinitely.
			if err := p.packetConn.SetReadDeadline(time.Now().Add(time.Millisecond * 10)); p.closed(err) {
				close(p.wc)                      // Close write channel if connection is closed.
				p.trace("startRead() closed wc") // Log write channel closure.
				return
			}
			// Read packet data from the connection.
			n, srcAddr, err := p.packetConn.ReadFrom(buf)
			if p.closed(err) {
				close(p.wc)                      // Close write channel if connection is closed.
				p.trace("startRead() closed wc") // Log write channel closure.
				return
			}
			if n > 0 && srcAddr != nil {
				buf2 := buf[:n] // Slice buffer to actual data size.
				// Parse received ICMP message.
				if msg, _ := icmp.ParseMessage(1, buf2); msg != nil {
					// Process the parsed message and send to write channel if valid.
					if pto := p.messageRead(msg, srcAddr); pto != nil {
						p.debug("conn->>>>>>ok: %s", pto.String()) // Log successful read.
						p.wc <- pto                                // Send Proto message to write channel.
					}
				}
			}
		}
	}
}

// messageRead processes received ICMP messages and returns a Proto instance if valid.
func (p *packet) messageRead(msg *icmp.Message, srcAddr net.Addr) (pto *Proto) {
	// parseEcho processes ICMP Echo Reply messages and constructs a Proto instance.
	parseEcho := func(ec *icmp.Echo) (pto *Proto) {
		if ec != nil && ec.ID > 0 {
			// Retrieve TTL and RTT for the echo message.
			if ttl, rtt := p.getTTL(ec); rtt > 0 {
				pto = pongProto(ttl, ec.ID, ec.Seq, srcAddr, aip4(srcAddr), rtt) // Create Proto instance.
			}
		}
		return
	}

	switch msg.Type {
	case ipv4.ICMPTypeEchoReply:
		// Handle ICMP Echo Reply messages.
		return parseEcho(msg.Body.(*icmp.Echo))

	case ipv4.ICMPTypeTimeExceeded:
		// Handle ICMP Time Exceeded messages (e.g., TTL expired).
		ee, ok := msg.Body.(*icmp.TimeExceeded)
		if !ok {
			return // Return nil if body is not TimeExceeded.
		}
		// Parse the original message embedded in the Time Exceeded message.
		msg0, _ := icmp.ParseMessage(1, ee.Data[20:])
		if msg0 == nil {
			return // Return nil if parsing fails.
		}
		msgBody := msg0.Body
		if msgBody == nil {
			return // Return nil if body is missing.
		}
		// Process the embedded Echo message.
		return parseEcho(msgBody.(*icmp.Echo))
	}
	return // Return nil for unhandled message types.
}

// setTTL stores TTL and timestamp information for a packet in the map.
func (p *packet) setTTL(ttl, id, seq int) {
	p.mu.Lock()                        // Lock for thread-safe map access.
	defer p.mu.Unlock()                // Unlock after map access.
	k := fmt.Sprintf("%d-%d", id, seq) // Create key from ID and sequence number.
	now := time.Now().UnixMilli()      // Get current timestamp.
	p.m[k] = ttlOpt{ttl, now}          // Store TTL and timestamp.
}

// getTTL retrieves TTL and calculates round-trip time (RTT) for a packet.
func (p *packet) getTTL(ec *icmp.Echo) (ttl int, rtt time.Duration) {
	p.mu.Lock()                              // Lock for thread-safe map access.
	defer p.mu.Unlock()                      // Unlock after map access.
	k := fmt.Sprintf("%d-%d", ec.ID, ec.Seq) // Create key from ID and sequence number.
	opt, ok := p.m[k]                        // Retrieve TTL option from map.
	if !ok {
		return // Return zero values if not found.
	}
	delete(p.m, k)                // Remove entry from map.
	now := time.Now().UnixMilli() // Get current timestamp.
	ms := now - opt.unix          // Calculate time difference in milliseconds.
	if ms == 0 {
		ms = 1 // Ensure non-zero RTT.
	}
	return opt.ttl, time.Duration(ms) * time.Millisecond // Return TTL and RTT.
}

// closed checks if an error indicates a closed network connection.
func (p *packet) closed(err error) (closed bool) {
	return err != nil && strings.HasSuffix(err.Error(), "use of closed network connection")
}
