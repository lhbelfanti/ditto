package http

import (
	"context"
	"io"

	"github.com/lhbelfanti/ditto/v2/log"
)

// closeBody closes body, and logs a failure at Warn: the response was already read.
func closeBody(ctx context.Context, body io.ReadCloser) {
	err := body.Close()
	if err != nil {
		log.Warn(log.With(ctx, log.Param("cause", err)), "failed to close the response body")
	}
}
