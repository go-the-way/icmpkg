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

// Package icmpkg provides functionality for performing ICMP-based ping and traceroute operations.
// It supports sending and receiving ICMP Echo Request and Reply messages, handling TTL (Time To Live)
// configurations, and processing responses with round-trip time (RTT) calculations. The package is
// designed to be flexible, supporting both ping and traceroute modes with configurable parameters
// such as TTL, packet count, and timeouts.
//
// The package includes the following main components:
//   - Proto: Represents an ICMP packet's metadata, including TTL, ID, sequence number, address, and RTT.
//   - packet: Manages low-level ICMP packet sending and receiving, with support for concurrent read/write operations.
//   - traceroute: Implements ping and traceroute functionality, handling multiple TTLs, packet sequences, and response processing.
//   - Ping and Traceroute functions: High-level interfaces for initiating ping or traceroute operations with customizable durations.
//
// Key features include:
//   - Support for both ping and traceroute modes, distinguished by the traceroute flag.
//   - Configurable write and read timeouts for flexible operation timing.
//   - Thread-safe handling of ICMP packets using mutexes and atomic operations.
//   - Debug and trace logging controlled by environment variables (e.g., ICMPKT_DEBUG, PING_DEBUG, TRACEROUTE_DEBUG).
//   - Context support for operation cancellation.
//   - Customizable pong handlers for processing ICMP responses.
//
// Usage examples:
//
//	// Ping example: Perform a ping to 8.8.8.8 with 3 packets.
//	ping := icmpkg.Ping("8.8.8.8", 3)
//	ping.PongHandler(func(pong *icmpkg.Proto) {
//	    fmt.Printf("Received: %s\n", pong.String())
//	})
//	ping.Run()
//
//	// Traceroute example: Perform a traceroute to 8.8.8.8 with max TTL 30 and 3 packets per TTL.
//	tr := icmpkg.Traceroute("8.8.8.8", 30, 3)
//	tr.PongHandler(func(pong *icmpkg.Proto) {
//	    fmt.Printf("Received: %s\n", pong.String())
//	})
//	tr.Run()
//
// Environment variables:
//   - ICMPKT_DEBUG: Set to "T" to enable debug logging for packet operations.
//   - ICMPKT_TRACE: Set to "T" to enable trace logging for packet operations.
//   - PING_DEBUG: Set to "T" to enable debug logging for ping operations.
//   - PING_TRACE: Set to "T" to enable trace logging for ping operations.
//   - TRACEROUTE_DEBUG: Set to "T" to enable debug logging for traceroute operations.
//   - TRACEROUTE_TRACE: Set to "T" to enable trace logging for traceroute operations.
//
// The package uses the "golang.org/x/net/icmp" and "golang.org/x/net/ipv4" packages for low-level
// ICMP communication and is designed to work with IPv4 networks.
package icmpkg
