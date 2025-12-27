# cold2warm

> **Note:** This tool was created to explore Go concurrency patterns with goroutines. For production use cases, you might prefer established tools like [s3cmd](https://s3tools.org/s3cmd) or [AWS CLI](https://aws.amazon.com/cli/), which have extensive documentation and community support available online.

A CLI tool to bulk-restore S3 objects from archival storage classes using concurrent goroutines.

## Features

- Uses Go goroutines to request glacier object restore operations
- Supports dry run mode for testing
- Displays comprehensive summary with the following metrics (shown in structured logs):
  | Metric | Description | Available in Dry Run |
  |--------|-------------|---------------------|
  | `total_objects_count` | Total number of Glacier objects found in the bucket | Yes |
  | `total_objects_size` | Combined size of all Glacier objects (human-readable format) | Yes |
  | `avg_obj_size` | Average size per Glacier object (human-readable format) | Yes |
  | `elapsed_time` | Elapsed time | Yes |
  | `inprogress_restore_count` | Number of objects with `RestoreAlreadyInProgress` status | No |
  | `total_inprogress_object_size` | Combined size of objects already being restored (human-readable format) | No |
  | `failed_restore_count` | Number of failed restore requests (excluding in-progress) | No |

> **Note:** All metrics are displayed in structured log format as key-value pairs in the final summary message.

## Usage

```bash
cold2warm --help
Usage: cold2warm --s3-endpoint=STRING --s3-access-key=STRING --s3-secret-key=STRING --s3-bucket-name=STRING [flags]

A CLI tool to bulk-restore S3 objects from archival storage classes using concurrent goroutines.

Flags:
  -h, --help                     Show context-sensitive help.
      --worker-count=10          Number of worker goroutines ($WROKER_COUNT)
      --dry-run                  Simulate operations without actually restoring objects ($DRY_RUN)
      --s3-region="nl-ams"       The region where the S3 bucket is hosted (e.g., nl-ams) ($S3_REGION).
      --s3-endpoint=STRING       Custom S3 endpoint URL (e.g., s3.nl-ams.scw.cloud). Do NOT include the bucket name ($S3_ENDPOINT).
      --s3-access-key=STRING     The access key ID for S3 authentication ($S3_ACCESS_KEY).
      --s3-secret-key=STRING     The secret access key for S3 authentication ($S3_SECRET_KEY).
      --s3-bucket-name=STRING    The name of the target S3 bucket ($S3_BUCKET_NAME).
      --s3-days=30               Number of days to keep the restored object ($S3_RESTORE_DAYS).
      --s3-prefix=""             Filter objects by this prefix (e.g., 'backups/') ($S3_OBJECT_PREFIX).
      --log-format="json"        Set the output format of the logs. Must be "console" or "json" ($LOG_FORMAT).
      --log-level=INFO           Set the log level. Must be "DEBUG", "INFO", "WARN" or "ERROR" ($LOG_LEVEL).
      --log-add-source           Whether to add source file and line number to log records ($LOG_ADD_SOURCE).
```

Example output

```bash
task run
time=2025-12-27T14:09:56+01:00 level=INFO msg="Version information" version="" branch="" revision=""
time=2025-12-27T14:09:56+01:00 level=INFO msg="Build context" go_version=go1.25.5 user="" date=""
time=2025-12-27T14:09:56+01:00 level=INFO msg="Starting Glacier object restoration" workers=20 bucket=REDACTED region=REDACTED prefix=immich restore_duration_days=3 dry_run=false
time=2025-12-27T14:10:16+01:00 level=INFO msg=Summary avg_obj_size="23 MB" failed_restore_count=0 inprogress_restore_count=9528 elapsed_time=20s total_inprogress_object_size="220 GB" total_objects_count=9528 total_objects_size="220 GB"
```

## Installation

Homebrew

```bash
brew install --cask veerendra2/tap/cold2warm
```

Download binaries

```bash
LATEST_VERSION=$(curl -s https://api.github.com/repos/veerendra2/cold2warm/releases/latest | jq -r '.tag_name')
curl -sL -o /tmp/cold2warm.tar.gz https://github.com/veerendra2/cold2warm/releases/download/${LATEST_VERSION}/cold2warm_$(uname -s | tr '[:upper:]' '[:lower:]')_$(arch).zip
unzip /tmp/cold2warm.zip -d /tmp
chmod +x /tmp/cold2warm
sudo mv /tmp/cold2warm /usr/local/bin/
```

## Local Development

```bash
# Deploy localstack (https://github.com/localstack/localstack)
docker compose -f compose-dev.yml up -d

# Run the app (https://taskfile.dev/)
task run

# Build the application
task build

# List available tasks
task --list
task: Available tasks for this project:
* all:                   Run comprehensive checks: format, lint, security and test
* build:                 Build the application binary for the current platform
* build-docker:          Build Docker image
* build-platforms:       Build the application binaries for multiple platforms and architectures
* fmt:                   Formats all Go source files
* install:               Install required tools and dependencies
* lint:                  Run static analysis and code linting using golangci-lint
* run:                   Runs the main application
* security:              Run security vulnerability scan
* test:                  Runs all tests in the project      (aliases: tests)
* vet:                   Examines Go source code and reports suspicious constructs
```

## References

- [aws-doc-sdk-examples](https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/go/example_code/s3)
- [Amazon S3 ListObjectsV2 Example](https://github.com/aws/aws-sdk-go-v2/tree/main/example/service/s3/listObjects)
- [Configure Client Endpoints](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-endpoints.html)
- [Configure the SDK](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html)
- [Performing Basic Amazon S3 Bucket Operations](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/s3-example-basic-bucket-operations.html)
