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
	"encoding/json"
	"log"
	"strings"
	"time"
)

var monitorBenchmark time.Duration

func logActivity(body []byte) {
	var logs Log
	// Read configuration
	err := json.Unmarshal(body, &logs)
	if err != nil {
		log.Println(err)
		return
	}

	logBucket := strings.Split(logs.Key, "/")[0]

	for idBucket, bucket := range configBuckets {
		for idPrefix, prefix := range bucket.Prefixes {
			if logBucket == bucket.Name && strings.Contains(logs.Key, logBucket+"/"+prefix.Name) {
				log.Printf("noop - found %s activity on bucket %v prefix %v - resetting alert...\n", logs.EventName, bucket.Name, prefix.Name)
				for _, logEntry := range logs.Records {
					t, err := time.Parse("2006-01-02T15:04:05Z", logEntry.EventTime)
					if err != nil {
						log.Println(err)
						continue
					}
					configBuckets[idBucket].Prefixes[idPrefix].ActivityTime = t.Format("2006-01-02T15:04:05Z")
				}
			}
		}
	}
}

// monitorSchedule begins the monitor schedule
func monitorSchedule(doneCh <-chan struct{}) {
	ticker := time.NewTicker(time.Second * time.Duration(MonitorInterval))
	go func() {
		for {
			select {
			case <-ticker.C:
				start := time.Now()
				monitorActivity()
				monitorBenchmark = time.Since(start)
			case <-doneCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func monitorActivity() {
	now := time.Now().UTC()
	for idBucket, bucket := range configBuckets {
		for idPrefix, prefix := range bucket.Prefixes {
			logActivityTime := configBuckets[idBucket].Prefixes[idPrefix].ActivityTime
			if logActivityTime == "" {
				configBuckets[idBucket].Prefixes[idPrefix].ActivityTime = now.Format("2006-01-02T15:04:05Z")
				logActivityTime = configBuckets[idBucket].Prefixes[idPrefix].ActivityTime
			}
			activityTime, err := time.Parse("2006-01-02T15:04:05Z", logActivityTime)
			if err != nil {
				log.Println(err)
				continue
			}

			timeThreshold := now.Add(time.Duration(-alertThreshold) * time.Second)
			cooldownThreshold := now.Add(time.Duration(-MonitorCoolDown) * time.Second)

			if activityTime.Before(timeThreshold) {
				firstFire := false
				logAlertFireTime := configBuckets[idBucket].Prefixes[idPrefix].AlertFireTime
				if logAlertFireTime == "" {
					firstFire = true
					configBuckets[idBucket].Prefixes[idPrefix].AlertFireTime = now.Format("2006-01-02T15:04:05Z")
					logAlertFireTime = configBuckets[idBucket].Prefixes[idPrefix].AlertFireTime
				}
				alertFireTime, err := time.Parse("2006-01-02T15:04:05Z", logAlertFireTime)
				if err != nil {
					log.Println(err)
					continue
				}
				if firstFire || alertFireTime.Before(cooldownThreshold) {
					configBuckets[idBucket].Prefixes[idPrefix].AlertFireTime = now.Format("2006-01-02T15:04:05Z")
					err = alert(bucket.Name, prefix.Name, activityTime)
					if err != nil {
						log.Println(err)
					}
					log.Printf("cooling down for %vs\n", MonitorCoolDown)
					log.Printf("last scan: %v\n", monitorBenchmark)

				}
			}
		}
	}
}
