// Copyright 2025 icmpkg Author. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package icmpkg

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Proto represents an ICMP packet's metadata, including TTL, identifiers, and timing information.
type Proto struct {
	TTL  int           // Time To Live value for the packet.
	ID   int           // Identifier for the ICMP packet.
	Seq  int           // Sequence number for the ICMP packet.
	Addr net.Addr      // Network address of the destination or source.
	Ip4  string        // IPv4 address as a string.
	Rtt  time.Duration // Round-trip time for the packet.
}

// pingProto creates a Proto instance for an ICMP Echo Request (ping).
func pingProto(ttl, id, seq int, addr net.Addr, ip4 string) *Proto {
	// Initialize a Proto instance with the provided TTL, ID, sequence number, address, and IPv4 string.
	return &Proto{TTL: ttl, ID: id, Seq: seq, Addr: addr, Ip4: ip4}
}

// pongProto creates a Proto instance for an ICMP Echo Reply (pong) with round-trip time.
func pongProto(ttl, id, seq int, addr net.Addr, ip4 string, rtt time.Duration) *Proto {
	// Initialize a Proto instance with the provided TTL, ID, sequence number, address, IPv4 string, and round-trip time.
	return &Proto{TTL: ttl, ID: id, Seq: seq, Addr: addr, Ip4: ip4, Rtt: rtt}
}

// timeoutProto creates a Proto instance for an ICMP timeout event (e.g., TTL exceeded).
func timeoutProto(ttl, id, seq int) *Proto {
	// Initialize a Proto instance with the provided TTL, ID, and sequence number, leaving other fields empty.
	return &Proto{TTL: ttl, ID: id, Seq: seq}
}

// String returns a string representation of the Proto instance for logging or debugging.
func (p *Proto) String() string {
	// Format the Proto fields into a human-readable string.
	return fmt.Sprintf("TTL: %d, ID: %d, Seq: %d, Addr: %v, Ip4: %v, Rtt: %v", p.TTL, p.ID, p.Seq, p.Addr, p.Ip4, p.Rtt)
}

// buf generates the byte representation of an ICMP Echo Request message for the Proto instance.
func (p *Proto) buf() []byte {
	// Create an ICMP Echo Request message with the Proto's ID and sequence number.
	msg := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Body: &icmp.Echo{
			ID:  p.ID,
			Seq: p.Seq,
		},
	}
	// Marshal the message into a byte slice, ignoring any errors.
	buf, _ := msg.Marshal(nil)
	return buf
}
