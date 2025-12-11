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
      --s3-days=30               Number of days to keep the restored object ($S3_RESTORE_DAYS)
      --s3-prefix=""             Filter objects by this prefix (e.g., 'backups/') ($S3_OBJECT_PREFIX).
      --worker-count=10          Number of worker goroutines ($WORKER_COUNT)
      --log-format="json"        Set the output format of the logs. Must be "console" or "json" ($LOG_FORMAT).
      --log-level=INFO           Set the log level. Must be "DEBUG", "INFO", "WARN" or "ERROR" ($LOG_LEVEL).
      --log-add-source           Whether to add source file and line number to log records ($LOG_ADD_SOURCE).
```

## Local Development

### Test Locally

Start [localstack](https://github.com/localstack/localstack)

```bash
docker compose -f compose-dev.yml up
```

Start program

```bash
task run
```

### Build & Test

- Using [Taskfile](https://taskfile.dev/)

_Install Taskfile: [Installation Guide](https://taskfile.dev/docs/installation)_

```bash
# List available tasks
task --list
task: Available tasks for this project:
* build:                 Build the application binary for the current platform
* build-platforms:       Build the application binaries for multiple platforms and architectures
* fmt:                   Formats all Go source files
* run:                   Runs the main application
* test:                  Runs all tests in the project      (aliases: tests)
* vet:                   Examines Go source code and reports suspicious constructs

# Build the application
task build

# Run tests
task test
```

- Build with [goreleaser](https://goreleaser.com/)

_Install GoReleaser: [Installation Guide](https://goreleaser.com/install/)_

```bash
# Build locally
goreleaser release --snapshot --clean
...
```

## References

- [aws-doc-sdk-examples](https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/go/example_code/s3)
- [Amazon S3 ListObjectsV2 Example](https://github.com/aws/aws-sdk-go-v2/tree/main/example/service/s3/listObjects)
- [Configure Client Endpoints](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-endpoints.html)
- [Configure the SDK](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html)
- [Performing Basic Amazon S3 Bucket Operations](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/s3-example-basic-bucket-operations.html)
