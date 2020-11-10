package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dekobon/distro-detect/env"
	"github.com/dekobon/distro-detect/linux"
	"log"
	"os"
	"strings"
)

func main() {
	var format string
	var fields string
	var fsRoot string

	flag.StringVar(&format, "format", "text", "Output format - valid values: text, text-no-labels, json, json-one-line")
	flag.StringVar(&fields, "fields", "", "Fields to output (comma separated) - valid values: name, id, version, lsb_release, os_release")
	flag.StringVar(&fsRoot, "fsroot", "/", "Path to the root of the filesystem in which to detect distro")

	flag.Parse()

	logger := log.New(os.Stderr, "error: ", 0)

	linux.FileSystemRoot = fsRoot
	distro := linux.DiscoverDistro()

	// Plain text output
	if format == "text" || format == "text-no-labels" {
		var labelFormat string
		if format == "text" {
			labelFormat = "%s: "
		} else if format == "text-no-labels" {
			labelFormat = ""
		}

		if fields == "" {
			err := distro.WriteAllResults(labelFormat, os.Stdout)
			if err != nil {
				logger.Println(err)
				os.Exit(-1)
			}
		} else {
			distroDetails := distro.AsMap()
			segments := strings.Split(fields, ",")
			for i := 0; i < len(segments); i++ {
				key := strings.ToLower(strings.TrimSpace(segments[i]))

				if distroDetails[segments[i]] != "" {
					err := distro.WriteResult(labelFormat, key, os.Stdout)
					if err != nil {
						logger.Println(err)
						os.Exit(-1)
					}
				}
			}
		}

		os.Exit(0)
	}

	// JSON output
	if format == "json" || format == "json-one-line" {
		var jsonOutput []byte
		var err error

		if format == "json" {
			jsonOutput, err = json.MarshalIndent(distro, "", "  ")
		} else if format == "json-one-line" {
			jsonOutput, err = json.Marshal(distro)
		}

		if err != nil {
			logger.Println(err)
			os.Exit(-1)
		}

		fmt.Printf("%s%s", jsonOutput, env.LineBreak)
		os.Exit(0)
	}
}
