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

package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"time"

	"github.com/go-the-way/icmpkg"
	"github.com/spf13/cobra"
)

// rootCmd represents the goping root command
var rootCmd = &cobra.Command{
	Use:   "goping [target]",
	Short: "goping is a command-line tool for ICMP ping",
	Long: `goping is a command-line tool based on the icmpkg package for performing ICMP ping operations.
It supports configuration of target address, packet count, write timeout, read timeout, packet ID, sequence number,
output format (text, json, xml), and signal handling for graceful shutdown.`,
	Args: cobra.ExactArgs(1), // Requires exactly one argument (target address)
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set debug and trace environment variables
		if debug {
			os.Setenv("ICMPKG_DEBUG", "T")
			os.Setenv("PING_DEBUG", "T")
		}
		if trace {
			os.Setenv("ICMPKG_TRACE", "T")
			os.Setenv("PING_TRACE", "T")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		ping := icmpkg.PingDuration(target, count, writeTimeout, readTimeout)
		var stats pingStats
		sys := !textOutput && !jsonOutput && !xmlOutput
		if sys {
			// Print header similar to system ping
			fmt.Printf("PING %s (%s) 56 bytes of data.\n", target, ping.Ip4())
		}

		// Set PongHandler based on output format
		ping.PongHandler(func(pong *icmpkg.Proto) {
			outputProto := protoOutput{
				ID:  pong.ID,
				Seq: pong.Seq,
				Ip4: pong.Ip4,
				Rtt: pong.Rtt,
			}
			if jsonOutput {
				data, _ := json.Marshal(outputProto)
				fmt.Println(string(data))
			} else if xmlOutput {
				data, _ := xml.Marshal(outputProto)
				fmt.Printf("%s\n", data)
			} else if textOutput {
				fmt.Println(outputProto.String())
			} else {
				// System ping-style output
				stats.transmitted++
				if pong.Rtt == 0 {
					fmt.Printf("Request timeout for icmp_id %d icmp_seq %d\n", pong.ID, pong.Seq)
				} else {
					stats.received++
					fmt.Printf("64 bytes from %s: icmp_id=%d icmp_seq=%d time=%d ms\n", pong.Ip4, pong.ID, pong.Seq, pong.Rtt.Milliseconds())
				}
				rttMs := float64(pong.Rtt) / float64(time.Millisecond)
				stats.rttS = append(stats.rttS, rttMs)
			}
		})
		ping.Run()
		if sys {
			loss := float64(stats.transmitted-stats.received) / float64(stats.transmitted) * 100
			fmt.Printf("\n--- %s ping statistics ---\n", target)
			fmt.Printf("%d packets transmitted, %d received, %.1f%% packet loss\n", stats.transmitted, stats.received, loss)
			if len(stats.rttS) > 0 {
				min, avg, max, mdev := calculateRTTStats(stats.rttS)
				fmt.Printf("rtt min/avg/max/mdev = %.3f/%.3f/%.3f/%.3f ms\n", min, avg, max, mdev)
			}
		}
	},
}

// Command-line flags
var (
	count        int           // Number of ICMP packets to send
	writeTimeout time.Duration // Write timeout duration
	readTimeout  time.Duration // Read timeout duration
	textOutput   bool          // Enable Text output
	jsonOutput   bool          // Enable JSON output
	xmlOutput    bool          // Enable XML output
	debug        bool          // Enable debug logging
	trace        bool          // Enable trace logging
)

func init() {
	// Add flags
	rootCmd.Flags().IntVarP(&count, "count", "c", 3, "Number of ICMP packets to send")
	rootCmd.Flags().DurationVarP(&writeTimeout, "write-timeout", "w", 500*time.Millisecond, "Write timeout duration")
	rootCmd.Flags().DurationVarP(&readTimeout, "read-timeout", "r", 500*time.Millisecond, "Read timeout duration")
	rootCmd.Flags().BoolVarP(&textOutput, "text", "t", false, "Enable Text output")
	rootCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Enable JSON output")
	rootCmd.Flags().BoolVarP(&xmlOutput, "xml", "x", false, "Enable XML output")
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
