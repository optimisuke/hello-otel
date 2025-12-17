package telemetry

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	olog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger and OTLP log emitter.
type Logger struct {
	zap   *zap.Logger
	otel  olog.Logger
	level olog.Severity
}

// InitLoggerProvider builds an OTLP log exporter and zap logger.
// It respects standard OTEL_* env vars for endpoints.
func InitLoggerProvider(ctx context.Context) (*log.LoggerProvider, Logger, error) {
	exp, err := otlploggrpc.New(ctx)
	if err != nil {
		return nil, Logger{}, err
	}

	svcName := os.Getenv("OTEL_SERVICE_NAME")
	if svcName == "" {
		svcName = "todo-api-go"
	}

	res, err := resource.New(ctx,
		resource.WithTelemetrySDK(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithHost(),
		resource.WithFromEnv(),
		resource.WithAttributes(semconv.ServiceNameKey.String(svcName)),
	)
	if err != nil {
		return nil, Logger{}, err
	}

	lp := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(log.NewBatchProcessor(exp)),
	)

	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, Logger{}, err
	}

	return lp, Logger{zap: zapLogger, otel: lp.Logger("todo-api-go")}, nil
}

// Info logs with zap and OTLP.
func (l Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.emit(ctx, olog.SeverityInfo, msg, fields...)
	l.zap.With(fields...).Info(msg)
}

// Error logs with zap and OTLP.
func (l Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.emit(ctx, olog.SeverityError, msg, fields...)
	l.zap.With(fields...).Error(msg)
}

func (l Logger) emit(ctx context.Context, sev olog.Severity, msg string, fields ...zap.Field) {
	var rec olog.Record
	rec.SetObservedTimestamp(time.Now())
	rec.SetSeverity(sev)
	rec.SetBody(olog.StringValue(msg))

	for _, f := range fields {
		switch f.Type {
		case zapcore.StringType:
			rec.AddAttributes(olog.String(f.Key, f.String))
		case zapcore.BoolType:
			rec.AddAttributes(olog.Bool(f.Key, f.Integer == 1))
		case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
			rec.AddAttributes(olog.Int64(f.Key, f.Integer))
		case zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type, zapcore.UintptrType:
			rec.AddAttributes(olog.Int64(f.Key, int64(f.Integer)))
		case zapcore.Float64Type, zapcore.Float32Type:
			rec.AddAttributes(olog.Float64(f.Key, math.Float64frombits(uint64(f.Integer))))
		case zapcore.DurationType:
			rec.AddAttributes(olog.Int64(f.Key, f.Integer))
		case zapcore.TimeType:
			if f.Interface != nil {
				if t, ok := f.Interface.(time.Time); ok {
					rec.AddAttributes(olog.String(f.Key, t.Format(time.RFC3339Nano)))
				}
			}
		case zapcore.StringerType, zapcore.ReflectType:
			rec.AddAttributes(olog.String(f.Key, fmt.Sprint(f.Interface)))
		default:
			rec.AddAttributes(olog.String(f.Key, fmt.Sprint(f.Interface)))
		}
	}

	l.otel.Emit(ctx, rec)
}

// Sync flushes zap buffers.
func (l Logger) Sync() {
	_ = l.zap.Sync()
}
