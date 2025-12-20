# cold2warm

A CLI tool to bulk-restore S3 objects from archival storage classes using concurrent goroutines.

## Usage

```bash
cold2warm --help
Usage: cold2warm --s3-endpoint=STRING --s3-access-key=STRING --s3-secret-key=STRING --s3-bucket-name=STRING [flags]

A CLI tool to bulk-restore S3 objects from archival storage classes using concurrent goroutines.

Flags:
  -h, --help                     Show context-sensitive help.
      --s3-region="nl-ams"       The region where the S3 bucket is hosted (e.g., nl-ams) ($S3_REGION).
      --s3-endpoint=STRING       Custom S3 endpoint URL (e.g., s3.nl-ams.scw.cloud). Do NOT include the bucket name ($S3_ENDPOINT).
      --s3-access-key=STRING     The access key ID for S3 authentication ($S3_ACCESS_KEY).
      --s3-secret-key=STRING     The secret access key for S3 authentication ($S3_SECRET_KEY).
      --s3-bucket-name=STRING    The name of the target S3 bucket ($S3_BUCKET_NAME).
      --s3-days=30               Number of days to keep the restored object ($S3_RESTORE_DAYS).
      --s3-prefix=""             Filter objects by this prefix (e.g., 'backups/') ($S3_OBJECT_PREFIX).
      --worker-count=10          Number of worker goroutines ($WORKER_COUNT)
      --worker-dry-run           Simulate operations without actually restoring objects ($WORKER_DRY_RUN)
      --log-format="json"        Set the output format of the logs. Must be "console" or "json" ($LOG_FORMAT).
      --log-level=INFO           Set the log level. Must be "DEBUG", "INFO", "WARN" or "ERROR" ($LOG_LEVEL).
      --log-add-source           Whether to add source file and line number to log records ($LOG_ADD_SOURCE).
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
