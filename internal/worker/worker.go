package worker

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	humanize "github.com/dustin/go-humanize"
	"github.com/veerendra2/cold2warm/pkg/bucketmgr"
)

type Config struct {
	WorkersCount int  `name:"count" help:"Number of worker goroutines" env:"COUNT" default:"10"`
	DryRun       bool `name:"dry-run" help:"Simulate operations without actually restoring objects" env:"DRY_RUN" default:"false"`
}

type Summary struct {
	AvgObjectSize              int64
	FailedRestore              int64
	InProgressRestore          int64
	TotalInProgressObjectsSize int64
	TotalObjects               int64
	TotalObjectsSize           int64
}
type objectInfo struct {
	key  string
	size int64
}

func StreamObjects(ctx context.Context, p *s3.ListObjectsV2Paginator) <-chan objectInfo {
	objects := make(chan objectInfo, 32)

	go func() {
		defer close(objects)

		for p.HasMorePages() {
			pageCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			page, err := p.NextPage(pageCtx)
			cancel()
			if err != nil {
				slog.Error("Failed to get next page", "error", err)
				return
			}

			for _, obj := range page.Contents {
				// This ensures we exit quickly if the user cancels, even if
				// we are filtering many glacier storage class objects below.
				if ctx.Err() != nil {
					return
				}

				if obj.StorageClass != types.ObjectStorageClassGlacier {
					continue
				}

				objInfo := objectInfo{
					key:  *obj.Key,
					size: *obj.Size,
				}

				select {
				case <-ctx.Done():
					return
				case objects <- objInfo:
					slog.Debug("Found", "object", objInfo.key, "size", objInfo.size)
				}
			}
		}
	}()

	return objects
}

func Start(ctx context.Context, cfg Config, s3Client bucketmgr.Client) {
	var wg sync.WaitGroup
	var summary Summary

	if cfg.DryRun {
		slog.Info("DRY RUN MODE: No objects will actually be restored")
	}

	paginator := s3Client.ListObjectsPaginator(ctx)
	objChan := StreamObjects(ctx, paginator)

	wg.Add(cfg.WorkersCount)
	for range cfg.WorkersCount {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					slog.Info("Worker stopped due to cancellation")
					return
				case obj, ok := <-objChan:
					if !ok {
						slog.Debug("Worker finished, channel closed")
						return
					}

					if cfg.DryRun {
						slog.Debug("DRY RUN: Would restore", "object", obj.key, "size", obj.size)
					} else {
						slog.Debug("Restoring", "object", obj.key, "size", obj.size)
						reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
						err := s3Client.RestoreObject(reqCtx, obj.key)
						cancel()

						if err != nil {
							if strings.Contains(err.Error(), "RestoreAlreadyInProgress") {
								atomic.AddInt64(&summary.InProgressRestore, 1)
								atomic.AddInt64(&summary.TotalInProgressObjectsSize, obj.size)
								slog.Debug("Restore already in progress", "object", obj.key, "size", obj.size)
							} else {
								atomic.AddInt64(&summary.FailedRestore, 1)
								slog.Warn("Failed to restore", "object", obj.key, "size", obj.size, "error", err)
							}
						}
					}
					atomic.AddInt64(&summary.TotalObjects, 1)
					atomic.AddInt64(&summary.TotalObjectsSize, obj.size)
				}
			}

		}()
	}
	wg.Wait()
	if summary.TotalObjects > 0 {
		summary.AvgObjectSize = summary.TotalObjectsSize / summary.TotalObjects
	}

	slog.Info("Summary",
		"avg_obj_size", humanize.Bytes(uint64(summary.AvgObjectSize)),
		"failed_restore_count", summary.FailedRestore,
		"inprogress_restore_count", summary.InProgressRestore,
		"total_inprogress_object_size", humanize.Bytes(uint64(summary.TotalInProgressObjectsSize)),
		"total_objects_count", summary.TotalObjects,
		"total_objects_size", humanize.Bytes(uint64(summary.TotalObjectsSize)),
	)
}
