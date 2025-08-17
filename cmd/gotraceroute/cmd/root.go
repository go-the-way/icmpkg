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

// rootCmd represents the gotraceroute root command
var rootCmd = &cobra.Command{
	Use:   "gotraceroute [target]",
	Short: "gotraceroute is a command-line tool for ICMP traceroute",
	Long: `gotraceroute is a command-line tool based on the icmpkg package for performing ICMP traceroute operations.
It supports configuration of target address, maximum TTL, packets per hop, write timeout, read timeout, packet ID,
sequence number, output format (text, json, xml), and signal handling for graceful shutdown.`,
	Args: cobra.ExactArgs(1), // Requires exactly one argument (target address)
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set debug and trace environment variables
		if debug {
			os.Setenv("ICMPKT_DEBUG", "T")
			os.Setenv("TRACEROUTE_DEBUG", "T")
		}
		if trace {
			os.Setenv("ICMPKT_TRACE", "T")
			os.Setenv("TRACEROUTE_TRACE", "T")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		tr := icmpkg.TracerouteDuration(target, maxTTL, count, writeTimeout, readTimeout)
		// Set PongHandler based on output format
		tr.PongHandler(func(pong *icmpkg.Proto) {
			outputProto := protoOutput{
				TTL: pong.TTL,
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
			} else {
				fmt.Println(pong.String())
			}
		})
		tr.Run()
	},
}

// Command-line flags
var (
	maxTTL       int           // Maximum TTL (hops)
	count        int           // Number of ICMP packets per hop
	writeTimeout time.Duration // Write timeout duration
	readTimeout  time.Duration // Read timeout duration
	jsonOutput   bool          // Enable JSON output
	xmlOutput    bool          // Enable XML output
	debug        bool          // Enable debug logging
	trace        bool          // Enable trace logging
)

func init() {
	// Add flags
	rootCmd.Flags().IntVarP(&maxTTL, "max-ttl", "m", 30, "Maximum TTL (hops)")
	rootCmd.Flags().IntVarP(&count, "count", "c", 3, "Number of ICMP packets per hop")
	rootCmd.Flags().DurationVarP(&writeTimeout, "write-timeout", "w", 500*time.Millisecond, "Write timeout duration")
	rootCmd.Flags().DurationVarP(&readTimeout, "read-timeout", "r", 500*time.Millisecond, "Read timeout duration")
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
