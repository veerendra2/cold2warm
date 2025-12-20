package worker

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/veerendra2/cold2warm/pkg/bucketmgr"
)

type Config struct {
	WorkersCount int  `name:"count" help:"Number of worker goroutines" env:"COUNT" default:"10"`
	DryRun       bool `name:"dry-run" help:"Simulate operations without actually restoring objects" env:"DRY_RUN" default:"false"`
}

func StreamObjects(ctx context.Context, p *s3.ListObjectsV2Paginator) <-chan string {
	objects := make(chan string, 32)

	go func() {
		defer close(objects)

		for p.HasMorePages() {
			pageCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			page, err := p.NextPage(pageCtx)
			cancel()
			if err != nil {
				slog.Error("failed to get next page", "error", err)
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

				select {
				case <-ctx.Done():
					return
				case objects <- *obj.Key:
					slog.Debug("Found", "object", *obj.Key)
				}
			}
		}
	}()

	return objects
}

func Start(ctx context.Context, cfg Config, s3Client bucketmgr.Client) {
	var wg sync.WaitGroup
	var totalObjects int64

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
						slog.Info("DRY RUN: Would restore object", "object", obj)
						atomic.AddInt64(&totalObjects, 1)
					} else {
						slog.Debug("Restoring", "object", obj)
						reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
						err := s3Client.RestoreObject(reqCtx, obj)
						cancel()
						if err != nil {
							slog.Warn("failed to restore", "object", obj, "error", err)
						} else {
							atomic.AddInt64(&totalObjects, 1)
						}
					}

				}
			}

		}()
	}
	wg.Wait()
	slog.Info("Total glacier objects restored", "count", atomic.LoadInt64(&totalObjects))
}
