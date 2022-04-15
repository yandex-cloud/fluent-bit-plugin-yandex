package plugin

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/util"

	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

func (p *Plugin) WriteAll(resourceToEntries map[Resource][]*logging.IncomingLogEntry) (results chan error, resCount int) {
	const batchMaxLen = 100
	resCount = 0
	for _, entries := range resourceToEntries {
		resCount += (len(entries) + batchMaxLen - 1) / batchMaxLen
	}
	results = make(chan error, resCount)

	for resource, entries := range resourceToEntries {
		resource := resource.LogEntryResource()
		entries := entries

		for len(entries) > 0 {
			var batch []*logging.IncomingLogEntry
			if len(entries) > batchMaxLen {
				batch, entries = entries[:batchMaxLen], entries[batchMaxLen:]
			} else {
				batch, entries = entries, nil
			}

			go func(res chan error) {
				err := p.write(context.Background(), batch, resource)
				res <- err
			}(results)
		}
	}

	return results, resCount
}

func (p *Plugin) write(ctx context.Context, entries []*logging.IncomingLogEntry, resource *logging.LogEntryResource) error {
	toSend := entries
	for len(toSend) > 0 {
		failed, err := p.client.Write(ctx, &logging.WriteRequest{
			Destination: p.destination,
			Resource:    resource,
			Entries:     toSend,
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
				toRetry = append(toRetry, toSend[idx])
			default:
				// bad message, just print
				fmt.Printf(
					"yc-logging: bad message %q: %q\n",
					util.Truncate(toSend[idx].GetMessage(), 512),
					failure.String(),
				)
			}
		}
		toSend = toRetry
	}
	return nil
}
