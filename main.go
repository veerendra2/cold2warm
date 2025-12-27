package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/dustin/go-humanize"
	"github.com/veerendra2/cold2warm/internal/worker"
	"github.com/veerendra2/cold2warm/pkg/bucketmgr"
	"github.com/veerendra2/gopackages/slogger"
	"github.com/veerendra2/gopackages/version"
)

const appName = "cold2warm"

var cli struct {
	Worker  worker.Config    `embed:""`
	S3      bucketmgr.Config `embed:"" prefix:"s3-" envprefix:"S3_"`
	Log     slogger.Config   `embed:"" prefix:"log-" envprefix:"LOG_"`
	Version kong.VersionFlag `name:"version" help:"Print version information and exit"`
}

func main() {
	kongCtx := kong.Parse(&cli,
		kong.Name(appName),
		kong.Description("A CLI tool to bulk-restore S3 objects from archival storage classes using concurrent goroutines."),
		kong.Vars{
			"version": version.Version,
		},
	)

	kongCtx.FatalIfErrorf(kongCtx.Error)

	slog.SetDefault(slogger.New(cli.Log))

	slog.Info("Version information", version.Info()...)
	slog.Info("Build context", version.BuildContext()...)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// Create s3 client
	s3Client, err := bucketmgr.NewClient(initCtx, cli.S3)
	if err != nil {
		slog.Error("Failed to create client", "error", err)
		kongCtx.Exit(1)
	}

	slog.Info("Starting Glacier object restoration",
		"workers", cli.Worker.WorkersCount,
		"bucket", cli.S3.BucketName,
		"region", cli.S3.Region,
		"prefix", cli.S3.ObjectPrefix,
		"restore_duration_days", cli.S3.Days,
		"dry_run", cli.Worker.DryRun,
	)

	startTime := time.Now()
	summary := worker.Start(ctx, cli.Worker, s3Client)
	if ctx.Err() == context.Canceled {
		slog.Error("Operation cancelled by user")
	}

	slog.Info("Summary",
		"avg_obj_size", humanize.Bytes(uint64(summary.AvgObjectSize)),
		"elapsed_time", time.Since(startTime).Round(time.Second).String(),
		"failed_restore_count", summary.FailedRestore,
		"inprogress_restore_count", summary.InProgressRestore,
		"total_inprogress_object_size", humanize.Bytes(uint64(summary.TotalInProgressObjectsSize)),
		"total_objects_count", summary.TotalObjects,
		"total_objects_size", humanize.Bytes(uint64(summary.TotalObjectsSize)),
	)
}
