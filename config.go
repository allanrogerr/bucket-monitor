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
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func setup() {
	// Read configuration
	s3Client, err := createClient(configEndpoint)
	if err != nil {
		log.Printf("Could not connect to s3 endpoint, searching on local filesystem %v\n", err)
	} else {
		// Fetch config
		err = fetchConfig(s3Client)
		if err != nil {
			log.Printf("Could not fetch config, searching on local filesystem %v\n", err)
		}
	}
	// Read config
	config, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Could not read config %v\n", err)
	}
	// Parse config
	err = json.Unmarshal(config, &configBuckets)
	if err != nil {
		log.Fatalf("Could not parse config %v\n", err)
	}
}

func createClient(configEndpoint string) (*minio.Client, error) {
	s3Client, err := minio.New(configEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(configAccessKey, configSecretKey, ""),
		Secure: true,
	})
	if err != nil {
		return nil, err
	}
	return s3Client, nil
}

func fetchConfig(s3Client *minio.Client) (err error) {
	if err := s3Client.FGetObject(context.Background(), configBucket, configFile, configFile, minio.GetObjectOptions{}); err != nil {
		return err
	}
	return nil
}
