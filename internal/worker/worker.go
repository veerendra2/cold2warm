package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/veerendra2/cold2warm/pkg/bucketmgr"
)

type Config struct {
	WorkersCount int `name:"count" help:"Number of worker goroutines" env:"COUNT" default:"10"`
}

func StreamObjects(ctx context.Context, p *s3.ListObjectsV2Paginator) (<-chan string, error) {
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

	return objects, nil
}

func Start(ctx context.Context, cfg Config, s3Client bucketmgr.Client) error {
	var wg sync.WaitGroup

	p, err := s3Client.ListObjectsPaginator(ctx)
	if err != nil {
		return err
	}

	objChan, err := StreamObjects(ctx, p)
	if err != nil {
		return err
	}

	wg.Add(cfg.WorkersCount)
	for range cfg.WorkersCount {
		go func() {
			defer wg.Done()

			for obj := range objChan {
				slog.Debug("Restoring", "object", obj)
				reqCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
				err := s3Client.RestoreObject(reqCtx, obj)
				cancel()
				if err != nil {
					slog.Warn("failed to restore", "object", obj, "error", err)
				}
			}
		}()
	}
	wg.Wait()

	return ctx.Err()
}
