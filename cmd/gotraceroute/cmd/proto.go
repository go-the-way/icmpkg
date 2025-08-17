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

package cmd

import (
	"fmt"
	"time"
)

// protoOutput adapts icmpkg.Proto for JSON/XML serialization
type protoOutput struct {
	TTL int           `json:"ttl" xml:"TTL"`
	ID  int           `json:"id" xml:"ID"`
	Seq int           `json:"seq" xml:"Seq"`
	Ip4 string        `json:"ip4" xml:"Ip4"`
	Rtt time.Duration `json:"rtt" xml:"Rtt"`
}

// String returns a string representation of the Proto instance for logging or debugging.
func (p *protoOutput) String() string {
	// Format the Proto fields into a human-readable string.
	return fmt.Sprintf("TTL: %d, ID: %d, Seq: %d, Ip4: %v, Rtt: %v", p.TTL, p.ID, p.Seq, p.Ip4, p.Rtt)
}
