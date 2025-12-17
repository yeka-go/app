package pgx

import (
	"context"
	"net"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type tracer struct {
	dbname string
}

func (t *tracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx, span := otel.Tracer("pgx").Start(ctx, "query")
	host, port, _ := net.SplitHostPort(conn.PgConn().Conn().RemoteAddr().String())
	span.SetAttributes(
		attribute.String("service_name", "pgx: "+t.dbname),
		attribute.String("db.name", t.dbname),
		attribute.String("db.query", data.SQL),
		attribute.String("server.address", host),
		attribute.String("server.port", port),
	)
	return ctx
}

func (t *tracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	if data.CommandTag.Insert() {
		span.SetName("INSERT")
	} else if data.CommandTag.Delete() {
		span.SetName("DELETE")
	} else if data.CommandTag.Select() {
		span.SetName("SELECT")
	} else if data.CommandTag.Update() {
		span.SetName("UPDATE")
	} else {
		span.SetName(data.CommandTag.String())
	}

	span.RecordError(data.Err)
	if data.Err != nil {
		span.SetStatus(codes.Error, data.Err.Error())
	}
}
