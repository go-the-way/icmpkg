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
	"net"
	"testing"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func TestPingProto(t *testing.T) {
	addr := &net.IPAddr{IP: net.ParseIP("8.8.8.8")}
	pto := pingProto(64, 1, 1, addr, "8.8.8.8")

	if pto == nil {
		t.Fatal("pingProto should return non-nil Proto")
	}
	if pto.TTL != 64 {
		t.Errorf("TTL = %d; want 64", pto.TTL)
	}
	if pto.ID != 1 {
		t.Errorf("ID = %d; want 1", pto.ID)
	}
	if pto.Seq != 1 {
		t.Errorf("Seq = %d; want 1", pto.Seq)
	}
	if pto.Addr != addr {
		t.Errorf("Addr = %v; want %v", pto.Addr, addr)
	}
	if pto.Ip4 != "8.8.8.8" {
		t.Errorf("Ip4 = %s; want 8.8.8.8", pto.Ip4)
	}
	if pto.Rtt != 0 {
		t.Errorf("Rtt = %v; want 0", pto.Rtt)
	}
}

func TestPongProto(t *testing.T) {
	addr := &net.IPAddr{IP: net.ParseIP("8.8.8.8")}
	rtt := time.Millisecond * 50
	pto := pongProto(64, 1, 1, addr, "8.8.8.8", rtt)

	if pto == nil {
		t.Fatal("pongProto should return non-nil Proto")
	}
	if pto.TTL != 64 {
		t.Errorf("TTL = %d; want 64", pto.TTL)
	}
	if pto.ID != 1 {
		t.Errorf("ID = %d; want 1", pto.ID)
	}
	if pto.Seq != 1 {
		t.Errorf("Seq = %d; want 1", pto.Seq)
	}
	if pto.Addr != addr {
		t.Errorf("Addr = %v; want %v", pto.Addr, addr)
	}
	if pto.Ip4 != "8.8.8.8" {
		t.Errorf("Ip4 = %s; want 8.8.8.8", pto.Ip4)
	}
	if pto.Rtt != rtt {
		t.Errorf("Rtt = %v; want %v", pto.Rtt, rtt)
	}
}

func TestTimeoutProto(t *testing.T) {
	pto := timeoutProto(64, 1, 1)

	if pto == nil {
		t.Fatal("timeoutProto should return non-nil Proto")
	}
	if pto.TTL != 64 {
		t.Errorf("TTL = %d; want 64", pto.TTL)
	}
	if pto.ID != 1 {
		t.Errorf("ID = %d; want 1", pto.ID)
	}
	if pto.Seq != 1 {
		t.Errorf("Seq = %d; want 1", pto.Seq)
	}
	if pto.Addr != nil {
		t.Errorf("Addr = %v; want nil", pto.Addr)
	}
	if pto.Ip4 != "" {
		t.Errorf("Ip4 = %s; want empty", pto.Ip4)
	}
	if pto.Rtt != 0 {
		t.Errorf("Rtt = %v; want 0", pto.Rtt)
	}
}

func TestProtoString(t *testing.T) {
	addr := &net.IPAddr{IP: net.ParseIP("8.8.8.8")}
	pto := &Proto{TTL: 64, ID: 1, Seq: 1, Addr: addr, Ip4: "8.8.8.8", Rtt: time.Millisecond * 50}
	expected := "{TTL: 64, ID: 1, Seq: 1, Addr: 8.8.8.8, Ip4: 8.8.8.8, Rtt: 50ms}"
	if got := pto.String(); got != expected {
		t.Errorf("String() = %q; want %q", got, expected)
	}
}

func TestProtoBuf(t *testing.T) {
	pto := &Proto{ID: 1, Seq: 1}
	buf := pto.buf()
	if len(buf) == 0 {
		t.Fatal("buf should return non-empty byte slice")
	}

	// Parse the buffer to verify it contains a valid ICMP Echo Request.
	msg, err := icmp.ParseMessage(1, buf)
	if err != nil {
		t.Fatalf("buf failed to parse: %v", err)
	}
	if msg.Type != ipv4.ICMPTypeEcho {
		t.Errorf("Message type = %v; want %v", msg.Type, ipv4.ICMPTypeEcho)
	}
	body, ok := msg.Body.(*icmp.Echo)
	if !ok {
		t.Fatal("Body is not ICMP Echo")
	}
	if body.ID != 1 {
		t.Errorf("ID = %d; want 1", body.ID)
	}
	if body.Seq != 1 {
		t.Errorf("Seq = %d; want 1", body.Seq)
	}
}
