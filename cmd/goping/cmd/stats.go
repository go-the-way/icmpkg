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
	"math"
	"sync"
)

// pingStats tracks statistics for system ping-style output
type pingStats struct {
	sync.Mutex
	transmitted int
	received    int
	rttS        []float64
}

// calculateRTTStats computes min, avg, max, and mdev for RTTs
func calculateRTTStats(rttS []float64) (min, avg, max, mdev float64) {
	if len(rttS) == 0 {
		return 0, 0, 0, 0
	}
	min = rttS[0]
	max = rttS[0]
	sum := 0.0
	for _, rtt := range rttS {
		if rtt < min {
			min = rtt
		}
		if rtt > max {
			max = rtt
		}
		sum += rtt
	}
	avg = sum / float64(len(rttS))

	// Calculate standard deviation (mdev)
	var sumSquaredDiff float64
	for _, rtt := range rttS {
		sumSquaredDiff += math.Pow(rtt-avg, 2)
	}
	mdev = math.Sqrt(sumSquaredDiff / float64(len(rttS)))
	return min, avg, max, mdev
}
