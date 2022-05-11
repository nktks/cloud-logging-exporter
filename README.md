# cloud-logging-exporter
`cloud-logging-exporter` is a command line agent tool to export file or stdin contents to Cloud Logging.

# Install
```
go install github.com/nakatamixi/cloud-logging-exporter/cmd/cloud-logging-exporter@latest
```

# Usage
```
./cloud-logging-exporter  -h
Usage of ./cloud-logging-exporter:
  -file string
    	target file path for export
  -logname string
    	log name for Cloud Logging (default "cloud-logging-exporter")
  -project string
    	GCP project
  -saved
    	export saved contents of file. (default true)
  -severity string
    	log severity(default|debug|info|notice|warning|error|critical|alert|emergency) (default "info")
  -touch
    	touch file if not exists.
```
You needs `GOOGLE_APPLICATION_CREDENTIALS` environment variables to authenticate Google Cloud.
And authenticated user needs `roles/logging.logWriter` role.

You can export stdin as below execution.
```
tail -f /path/to/file | ./cloud-logging-exporter
```
