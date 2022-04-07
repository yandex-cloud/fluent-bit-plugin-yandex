package plugin

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

func (p *Plugin) WriteAll(resourceToEntries map[Resource][]*logging.IncomingLogEntry) (results chan error, resBuffer int) {
	resBuffer = len(resourceToEntries)
	results = make(chan error, resBuffer)

	for resource, entries := range resourceToEntries {
		resource := resource.LogEntryResource()
		entries := entries

		go func(res chan error) {
			err := p.Write(context.Background(), entries, resource)
			res <- err
		}(results)
	}

	return results, resBuffer
}

func (p *Plugin) Write(ctx context.Context, entries []*logging.IncomingLogEntry, resource *logging.LogEntryResource) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	const batchMaxLen = 100
	for len(entries) > 0 {
		var batch []*logging.IncomingLogEntry
		if len(entries) > batchMaxLen {
			batch, entries = entries[:batchMaxLen], entries[batchMaxLen:]
		} else {
			batch, entries = entries, nil
		}
		failed, err := p.client.Write(ctx, &logging.WriteRequest{
			Destination: p.destination,
			Resource:    resource,
			Entries:     batch,
			Defaults:    p.defaults,
		})
		if err != nil {
			// return right away
			return err
		}
		var toRetry []*logging.IncomingLogEntry
		for idx, failure := range failed.GetErrors() {
			switch code := codes.Code(failure.GetCode()); code {
			case codes.ResourceExhausted,
				codes.FailedPrecondition,
				codes.Unavailable,
				codes.Unknown,
				codes.Canceled,
				codes.DeadlineExceeded:
				toRetry = append(toRetry, batch[idx])
			default:
				p.printMu.Lock()
				// bad message, just print
				fmt.Printf(
					"yc-logging: bad message %q: %q\n",
					truncate(batch[idx].GetMessage(), 512),
					failure.String(),
				)
				p.printMu.Unlock()
			}
		}
		// add back to retry later
		entries = append(toRetry, entries...)
	}
	return nil
}
