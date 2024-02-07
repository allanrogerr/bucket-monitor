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
	"log"
	"time"
)

func alert(bucketName string, prefixName string, activityTime time.Time) error {
	log.Printf("alert - no activity on %v/%v since %v\n", bucketName, prefixName, activityTime.Format("2006-01-02T15:04:05Z"))
	//from := env.Get("WEBHOOK_EMAIL_FROM", "")
	//pass := env.Get("WEBHOOK_EMAIL_PASS", "")
	//to := "mailing_list@example.com"
	//
	//msg := "From: " + from + "\n" +
	//	"To: " + to + "\n" +
	//	"Subject: alert - no activity on " + bucketName + "\n\n" +
	//	fmt.Sprintf("There has been no activity on %v/%v since %v\n", bucketName, prefixName, activityTime)
	//
	//err := smtp.SendMail("smtp.gmail.com:587",
	//	smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
	//	from, []string{to}, []byte(msg))
	//if err != nil {
	//	log.Printf("smtp error: %s", err)
	//	return err
	//}
	//log.Println("mail sent")
	return nil
}
