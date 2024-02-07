# Audit/Logger Webhook

`webhook` listens for incoming events from MinIO server and logs these events to a log file. This branch performs bucket and prefix activity monitoring

Usage:
```
./webhook --log-file <log-file> \
--threshold <threshold> \
--address <address> \
--config-file <config-file> \
--config-bucket <config-bucket> \
--config-endpoint <config-endpoint> \
--config-accesskey <config-accesskey> \
--config-secretkey <config-secretkey> 
```
e.g.
```
go build
./webhook --address 127.0.0.1:8080
```

This example comes with certain parameters baked in, assuming:
at s3 endpoint `play.min.io:9000` (`config-endpoint`), there exists bucket `config-store` (`config-bucket`) which contains object `config.json` (`config-file`) accessible using credentials `configreadonly` (`config-accesskey`) / `minio123` (`config-secretkey`)

The program logs listens on all interfaces at port `:8080` (`address`). Incoming events are all logged automatically to the file system at `./log.out` (`log-file`). By default, the program will check if no activity has been detected on a bucket for 15 min (`threshold`)

The following constants also control the program:
Activity scanning takes place every 5 seconds (`MonitorInterval`) and after a successful alert the system cools down for 15 s (`MonitorCoolDown`) before sending another alert if activity is still not detected on a specific bucket/prefix.

Example config-file. This needs to be loaded at `config-endpoint` in `config-bucket`
```
[
    {
        "name": "files0",
        "prefixes":
        [
            {
                "name": "demoA/Level0/Level1"
            },
            {
                "name": "demoB/Level3"
            }
        ]
    },
    {
        "name": "files1",
        "prefixes":
        [
            {
                "name": "demoX/Year2024/Month1/Day12"
            },
            {
                "name": "demoY/Year2023/Month6"
            },
            {
                "name": "demoX/Year2024/Month1/Day14"
            }
        ]
    }
]
```

Environment only settings:

| ENV                | Description                                                        |
|--------------------|--------------------------------------------------------------------|
| WEBHOOK_AUTH_TOKEN | Authorization token optional to authenticate/trust incoming events |

The webhook service can be setup as a systemd service using the `webhook.service` shipped with
this project

To send Audit logs from MinIO server, please configure MinIO using the command:
```
mc admin config set myminio audit_webhook endpoint=http://webhookendpoint:8080 auth_token=webhooksecret
```

To send server logs from MinIO server, please configure MinIO using the command:
```
mc admin config set myminio logger_webhook endpoint=http://webhookendpoint:8081 auth_token=webhooksecret
```

> NOTE: audit_webhook and logger_webhook should *not* be configured to send events to the same webhook instance.

Logs can be rotated using the standard logrotate tool. You can provide the postrotate command such that
webhook writes to a new log file after log rotation.
```
postrotate
	systemctl reload webhook
endscript
```
