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
	// MonitorInterval controls the frequency of activity monitoring (seconds)
	MonitorInterval = 5
	// MonitorCoolDown controls the frequency of alert sending, once an alert is active (seconds)
	MonitorCoolDown = 15
)

func main() {
	flag.StringVar(&logFile, "log-file", "log.out", "path to the file where webhook will log incoming events")
	flag.Int64Var(&alertThreshold, "threshold", 15*60, "maximum inactive time in bucket prefix")
	flag.StringVar(&address, "address", ":8080", "bind to a specific ADDRESS:PORT, ADDRESS can be an IP or hostname")

	flag.StringVar(&configFile, "config-file", "config.json", "path to the file from which bucket monitoring configurations will be read")
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

	l, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
	if err != nil {
		log.Fatal(err)
	}

	var mu sync.Mutex

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)

	go func() {
		for range sigs {
			mu.Lock()
			err := l.Sync()
			if err != nil {
				return
			} // flush to disk any temporary buffers.
			err = l.Close()
			if err != nil {
				return
			} // then close the file, before rotation.
			l, err = os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
			if err != nil {
				log.Fatal(err)
			}
			mu.Unlock()
		}
	}()

	setup()
	doneCh := make(chan struct{})
	monitorSchedule(doneCh)
	defer close(doneCh)

	go func() {
		log.Printf("Serving webhook at %v\n", address)
		err = http.ListenAndServe(address, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authToken != "" {
				if authToken != r.Header.Get("Authorization") {
					http.Error(w, "authorization header missing", http.StatusBadRequest)
					return
				}
			}
			switch r.Method {
			case http.MethodPost:
				mu.Lock()
				body, err := io.ReadAll(r.Body)
				if err != nil {
					mu.Unlock()
					return
				}
				logActivity(body)
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
				mu.Unlock()
			default:
			}
		}))
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-doneCh
}
