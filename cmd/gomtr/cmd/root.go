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
	"net"
	"os"
	"time"

	"github.com/go-the-way/icmpkg"
	"github.com/spf13/cobra"
)

// hop represents statistics for a single TTL hop
type hop struct {
	TTL                         int
	Addr                        string
	Sent, Received, Loss        int
	Sum, Last, Avg, Best, Worst int
}

func (h *hop) dataset(pong *icmpkg.Proto) {
	h.TTL = pong.TTL
	h.Sent++
	if h.Addr == "" && pong.Ip4 != "" {
		h.Addr = pong.Ip4
	}
	if pong.Rtt > 0 {
		h.Received++
		h.Last = int(pong.Rtt.Milliseconds())
		h.Sum += h.Last
		h.Best = max(min(h.Best, h.Last), h.Last)
		h.Worst = min(max(h.Worst, h.Last), h.Last)
		h.Avg = (h.Avg + h.Last) / 2
	}
	h.Loss = h.Received * 100 / h.Sent
	return
}

var hops [64]hop

func start() {
	tr := icmpkg.TracerouteDuration(target, maxTTL, count, interval, readTimeout)
	tr.PongHandler(pongHandler)

	prints(tr.Ip4())

	tr.Run()
}

func pongHandler(pong *icmpkg.Proto) {
	(&hops[pong.TTL]).dataset(pong)
}

// rootCmd represents the gomtr root command
var rootCmd = &cobra.Command{
	Use:   "gomtr [target]",
	Short: "gomtr is a command-line tool for ICMP-based MTR",
	Long: `gomtr is a command-line tool based on the icmpkg package for performing ICMP traceroute operations
with interactive terminal output similar to the mtr command. It supports configuration of target address,
maximum TTL, packets per hop, interval, read timeout, and debug/trace logging.`,
	Args: cobra.ExactArgs(1), // Requires exactly one argument (target address)
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set debug and trace environment variables
		if debug {
			os.Setenv("ICMPKG_DEBUG", "T")
			os.Setenv("MTR_DEBUG", "T")
		}
		if trace {
			os.Setenv("ICMPKG_TRACE", "T")
			os.Setenv("MTR_TRACE", "T")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		target = args[0]
		start()
	},
}

// Command-line flags
var (
	target      string
	maxTTL      int           // Maximum TTL (hops)
	count       int           // Number of ICMP packets per hop
	interval    time.Duration // Interval between packets
	readTimeout time.Duration // Read timeout duration
	debug       bool          // Enable debug logging
	trace       bool          // Enable trace logging
)

func init() {
	// Add flags
	rootCmd.Flags().IntVarP(&maxTTL, "max-ttl", "m", 30, "Maximum TTL (hops)")
	rootCmd.Flags().IntVarP(&count, "count", "c", 10, "Number of ICMP packets per hop")
	rootCmd.Flags().DurationVarP(&interval, "interval", "i", 100*time.Millisecond, "Interval between packets")
	rootCmd.Flags().DurationVarP(&readTimeout, "read-timeout", "r", 500*time.Millisecond, "Read timeout duration")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug logging")
	rootCmd.Flags().BoolVar(&trace, "trace", false, "Enable trace logging")
}

// Execute runs the root command
func Execute() {
	defer func() {
		if re := recover(); re != nil {
			fmt.Println(re)
		}
	}()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var hostname, _ = os.Hostname()

func localAddr() (addr string) {
	conn, _ := net.Dial("udp", target+":80")
	if conn != nil {
		addr = conn.LocalAddr().(*net.UDPAddr).IP.String()
		conn.Close()
	}
	return addr
}
