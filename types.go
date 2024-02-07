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

import "github.com/minio/madmin-go/v3"

// Log represents event information for some event targets.
type Log struct {
	EventName string
	Key       string
	Records   []Event
}

// Event represents event notification information defined in
// http://docs.aws.amazon.com/AmazonS3/latest/dev/notification-content-structure.html.
type Event struct {
	EventVersion      string            `json:"eventVersion"`
	EventSource       string            `json:"eventSource"`
	AwsRegion         string            `json:"awsRegion"`
	EventTime         string            `json:"eventTime"`
	EventName         string            `json:"eventName"`
	UserIdentity      Identity          `json:"userIdentity"`
	RequestParameters map[string]string `json:"requestParameters"`
	ResponseElements  map[string]string `json:"responseElements"`
	S3                Metadata          `json:"s3"`
	Source            Source            `json:"source"`
	Type              madmin.TraceType  `json:"-"`
}

// Name - event type enum.
// Refer http://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-event-types-and-destinations
// for most basic values we have since extend this and its not really much applicable other than a reference point.
// "s3:Replication:OperationCompletedReplication" is a MinIO extension.
type Name int

// Identity represents access key who caused the event.
type Identity struct {
	PrincipalID string `json:"principalId"`
}

// Metadata represents event metadata.
type Metadata struct {
	SchemaVersion   string `json:"s3SchemaVersion"`
	ConfigurationID string `json:"configurationId"`
	Bucket          Bucket `json:"bucket"`
	Object          Object `json:"object"`
}

// Bucket represents bucket metadata of the event.
type Bucket struct {
	Name          string   `json:"name"`
	OwnerIdentity Identity `json:"ownerIdentity"`
	ARN           string   `json:"arn"`
}

// Object represents object metadata of the event.
type Object struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size,omitempty"`
	ETag         string            `json:"eTag,omitempty"`
	ContentType  string            `json:"contentType,omitempty"`
	UserMetadata map[string]string `json:"userMetadata,omitempty"`
	VersionID    string            `json:"versionId,omitempty"`
	Sequencer    string            `json:"sequencer"`
}

// Source represents client information who triggered the event.
type Source struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	UserAgent string `json:"userAgent"`
}

// ConfigBucket is a custom type
type ConfigBucket struct {
	Name     string         `json:"name"`
	Prefixes []ConfigPrefix `json:"prefixes"`
}

// ConfigPrefix is a custom type
type ConfigPrefix struct {
	Name          string `json:"name"`
	ActivityTime  string `json:"activityTime,omitempty"`
	AlertFireTime string `json:"alertFireTime,omitempty"`
}
