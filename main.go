// Copyright (c) 2015-2023 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/minio/pkg/v2/env"
)

// Variables configured at program start from program parameters and other inputs
var (
	logFile         string
	configFile      string
	configEndpoint  string
	configAccessKey string
	configSecretKey string
	configBucket    string
	configBuckets   []ConfigBucket
	alertThreshold  int64
	address         string
	authToken       = env.Get("WEBHOOK_AUTH_TOKEN", "")
)

const (
	// MonitorInterval controls the frequency of activity monitoring (seconds) i.e. how often the monitorActivity function is called
	MonitorInterval = 5
	// MonitorCoolDown controls the frequency of alert sending, once an alert is active (seconds)
	MonitorCoolDown = 15
)

func main() {
	flag.StringVar(&logFile, "log-file", "log.out", "path to the file where webhook will log incoming events")
	flag.Int64Var(&alertThreshold, "threshold", 15*60, "maximum inactive time in bucket prefix")
	flag.StringVar(&address, "address", ":8080", "bind to a specific ADDRESS:PORT, ADDRESS can be an IP or hostname")

	flag.StringVar(&configFile, "config-file", "sample-config.json", "path to the file from which bucket monitoring configurations will be read")
	flag.StringVar(&configEndpoint, "config-endpoint", "play.min.io:9000", "s3 endpoint with config")
	flag.StringVar(&configAccessKey, "config-accesskey", "configreadonly", "access key of s3 endpoint with config")
	flag.StringVar(&configSecretKey, "config-secretkey", "minio123", "secret key of s3 endpoint with config")
	flag.StringVar(&configBucket, "config-bucket", "config-store", "bucket at s3 endpoint with config")

	flag.Parse()

	if logFile == "" {
		log.Fatalln("--log-file must be specified")
		return
	}

	if configFile == "" {
		log.Fatalln("--config-file must be specified")
		return
	}

	// logFile is opened with flags Create, Append, Write-Only and permissions 640
	l, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
	if err != nil {
		log.Fatal(err)
	}

	// Allows for mutual exclusion at the time of accessing variables, in this case
	var mu sync.Mutex

	// Channel accepting signals of reload, interrupt (ctrl+c) and terminate. These synchronous signals cause the program to exit
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for range sigs {
			mu.Lock()
			err := l.Sync()
			if err != nil {
				return
			} // Flush any temporary buffers to disk
			err = l.Close()
			if err != nil {
				return
			} // Close the file before rotation.
			l, err = os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
			if err != nil {
				log.Fatal(err)
			}
			mu.Unlock()
		}
	}()

	// Fetch json config from s3 endpoint, or from file system if the endpoint is inaccessible or the file cannot be found at the s3 endpoint
	setup()
	// Channel controlling graceful program exit
	doneCh := make(chan struct{})
	// Periodically (see `const MonitorInterval`) check the in-memory status of the config for alert candidates
	monitorSchedule(doneCh)
	defer close(doneCh)

	// Asynchronous function opening and listening to address:port (see var `address`) for Bucket Notifications
	go func() {
		log.Printf("Serving webhook at %v\n", address)
		err = http.ListenAndServe(address, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authToken != "" {
				if authToken != r.Header.Get("Authorization") {
					http.Error(w, "authorization header missing", http.StatusBadRequest)
					log.Println("authorization header missing")
					return
				}
			}
			switch r.Method {
			case http.MethodPost:
				// Locking objects to prevent other processes from modifying them
				mu.Lock()
				// Read notification payload
				body, err := io.ReadAll(r.Body)
				if err != nil {
					mu.Unlock()
					return
				}
				// Process payload
				logActivity(body)
				// Write to raw logfile
				_, err = io.Copy(l, bytes.NewReader(body))
				if err != nil {
					mu.Unlock()
					return
				}
				_, err = l.WriteString("\n")
				if err != nil {
					mu.Unlock()
					return
				}
				// Release lock
				mu.Unlock()
			default:
			}
		}))
		if err != nil {
			doneCh <- struct{}{}
			log.Fatal(err)
		}
	}()

	go func() {
		<-sigs
		doneCh <- struct{}{}
		log.Printf("Stopped serving webhook at %v\n", address)

	}()

	<-doneCh
}
