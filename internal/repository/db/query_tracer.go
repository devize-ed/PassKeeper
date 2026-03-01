package db

import (
	"context"
	"passKeper/internal/logger"

	"github.com/jackc/pgx/v5"
)

// queryTracer implements pgx.QueryTracer for logging query start/end.
type queryTracer struct{}

// TraceQueryStart logs the start of a query execution.
func (queryTracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceQueryStartData,
) context.Context {
	logger.Log.Debugf("Running query %s (%v)", data.SQL, data.Args)
	return ctx
}

// TraceQueryEnd logs the end of a query execution.
func (queryTracer) TraceQueryEnd(_ context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	logger.Log.Debugf("%v", data.CommandTag)
}
