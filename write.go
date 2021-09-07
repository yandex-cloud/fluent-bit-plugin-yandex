package main

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"google.golang.org/grpc/codes"
)

func (p *pluginImpl) write(ctx context.Context, entries []*logging.IncomingLogEntry) error {
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
			Resource:    p.resource,
			Entries:     batch,
			Defaults:    p.defaults,
		})
		if err != nil {
			// return right away
			return err
		}
		var toRetry []*logging.IncomingLogEntry
		for idx, entryToRetry := range failed.GetErrors() {
			switch code := codes.Code(entryToRetry.GetCode()); code {
			case codes.ResourceExhausted,
				codes.FailedPrecondition,
				codes.Unavailable,
				codes.Unknown,
				codes.Canceled,
				codes.DeadlineExceeded:
				toRetry = append(toRetry, batch[idx])
			default:
				// bad message, just print
				fmt.Printf("yc-logging: bad message %q: %q\n", batch[idx].String(), entryToRetry.String())
			}
		}
		// add back to retry later
		entries = append(toRetry, entries...)
	}
	return nil
}
