package plugin

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/model"

	"google.golang.org/grpc/codes"
)

func (p *Plugin) WriteAll(resourceToEntries map[model.Resource][]*model.Entry) (results chan error, resCount int) {
	const batchMaxLen = 100
	resCount = 0
	for _, entries := range resourceToEntries {
		resCount += (len(entries) + batchMaxLen - 1) / batchMaxLen
	}
	results = make(chan error, resCount)

	for resource, entries := range resourceToEntries {
		resource := resource
		entries := entries

		for len(entries) > 0 {
			var batch []*model.Entry
			if len(entries) > batchMaxLen {
				batch, entries = entries[:batchMaxLen], entries[batchMaxLen:]
			} else {
				batch, entries = entries, nil
			}

			go func(res chan error) {
				err := p.write(context.Background(), batch, &resource)
				res <- err
			}(results)
		}
	}

	return results, resCount
}

func (p *Plugin) write(ctx context.Context, entries []*model.Entry, resource *model.Resource) error {
	toSend := entries
	for len(toSend) > 0 {
		failed, err := p.client.Write(ctx, &model.WriteRequest{
			Resource: resource,
			Entries:  toSend,
		})
		if err != nil {
			// return right away
			return err
		}
		var toRetry []*model.Entry
		for idx, failure := range failed {
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
					truncate(toSend[idx].Message, 512),
					failure.String(),
				)
			}
		}
		toSend = toRetry
	}
	return nil
}
