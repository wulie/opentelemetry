package main

import (
	"context"
	"gitee.com/magusiiot/api/version"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"os"
)

var name = "version"

type VersionServer struct {
	version.UnimplementedMagusVersionServer
}

func (v *VersionServer) GetVersion(ctx context.Context, req *version.VersionReq) (*version.VersionReply, error) {
	var major, minor, patch int32
	major, ctx = GetMajor(ctx)
	//minor, ctx = GetMinor(ctx)
	//patch, ctx = GetPatch(ctx)

	return &version.VersionReply{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

//type MajorVersionNum int32
//type MinorVersionNum int32
//type PatchVersionNum int32
//
//type Major struct {
//	MajorVersionNum MajorVersionNum
//}

type mockTransport struct {
	kind      transport.Kind
	endpoint  string
	operation string
	//header    http.Header
}

func GetMajor(ctx context.Context) (int32, context.Context) {
	//svrTracer := tracing.NewTracer(trace.SpanKindServer, tracing.WithPropagator(propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})))
	//transporter, ok := transport.FromServerContext(ctx)
	//fmt.Println("ok.......",ok)
	//if ok {
	//	var span trace.Span
	//	ctx, span = svrTracer.Start(ctx, "GetMajor", transporter.RequestHeader())
	//	defer span.End()
	//}

	var minor int32

	minor, ctx = GetMinor(ctx)
	TranceSpan("GetMajor", ctx)

	return 1 + minor, ctx
}

func GetMinor(ctx context.Context) (int32, context.Context) {
	//svrTracer := tracing.NewTracer(trace.SpanKindServer, tracing.WithPropagator(propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})))
	//transporter, ok := transport.FromServerContext(ctx)
	//if ok {
	//	var span trace.Span
	//	ctx, span = svrTracer.Start(ctx, "GetMinor", transporter.RequestHeader())
	//	defer span.End()
	//}

	var patch int32
	patch, ctx = GetPatch(ctx)
	TranceSpan("GetMinor", ctx)

	return 33 + patch, ctx
}

func GetPatch(ctx context.Context) (int32, context.Context) {
	//svrTracer := tracing.NewTracer(trace.SpanKindServer, tracing.WithPropagator(propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})))
	//transporter, ok := transport.FromServerContext(ctx)
	//if ok {
	//	var span trace.Span
	//	ctx, span = svrTracer.Start(ctx, "GetPatch", transporter.RequestHeader())
	//	defer span.End()
	//}
	//_, span := otel.Tracer(name).Start(ctx, "GetPatch")
	//defer span.End()
	TranceSpan("GetPatch", ctx)
	return 45, ctx
}

func TranceSpan(funcName string, ctx context.Context) context.Context {
	svrTracer := tracing.NewTracer(trace.SpanKindServer, tracing.WithPropagator(propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})))
	transporter, ok := transport.FromServerContext(ctx)
	if ok {
		var span trace.Span
		ctx, span = svrTracer.Start(ctx, transporter.Operation(), transporter.RequestHeader())

		defer span.End()
	}
	return ctx
}

// set trace provider
func setTracerProvider(url string) error {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return err
	}
	tp := tracesdk.NewTracerProvider(
		// Set the sampling rate based on the parent span to 100%
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(1.0))),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String("version"),
			attribute.String("env", "dev"),
		)),
	)
	otel.SetTracerProvider(tp)
	return nil
}
func main() {
	logger := log.NewStdLogger(os.Stdout)
	logger = log.With(logger, "trace_id", tracing.TraceID())
	logger = log.With(logger, "span_id", tracing.SpanID())
	log := log.NewHelper(logger)

	url := "http://192.168.20.80:14268/api/traces"
	if os.Getenv("jaeger_url") != "" {
		url = os.Getenv("jaeger_url")
	}
	err := setTracerProvider(url)
	if err != nil {
		log.Error(err)
	}

	s := &VersionServer{}
	// grpc server
	grpcSrv := grpc.NewServer(
		grpc.Address(":9000"),
		grpc.Middleware(
			middleware.Chain(
				recovery.Recovery(),
				tracing.Server(),
				logging.Server(logger),
			),
		))
	//version.RegisterMessageServiceServer(grpcSrv, s)
	version.RegisterMagusVersionServer(grpcSrv, s)
	app := kratos.New(
		kratos.Name("version"),
		kratos.Server(
			grpcSrv,
		),
	)

	if err := app.Run(); err != nil {
		log.Error(err)
	}
}
