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
	"os"
	"time"
)

// Global variables controlling debug and trace logging based on environment variables.
var (
	pingDebug = os.Getenv("PING_DEBUG") == "T" // Enables debug logging if PING_DEBUG is set to "T".
	pingTrace = os.Getenv("PING_TRACE") == "T" // Enables trace logging if PING_TRACE is set to "T".
)

// ping is an alias for the traceroute type, used for ICMP ping operations.
type ping = traceroute

// Ping creates a ping instance with default write and read durations of 500ms.
func Ping(address string, count int) *ping {
	// Initialize ping with default durations for write and read operations.
	return PingDuration(address, count, time.Millisecond*500, time.Millisecond*500)
}

// PingDuration creates a ping instance with specified write and read durations.
func PingDuration(address string, count int, writeDur, readDur time.Duration) *ping {
	// Initialize a new traceroute instance for ping with the provided address, count, and durations.
	return newTraceroute(address, 1, count, writeDur, readDur, false)
}
