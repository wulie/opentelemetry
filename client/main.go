package main

import (
	"context"
	"fmt"
	"gitee.com/magusiiot/api/version"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"google.golang.org/grpc"
	"time"
)

const addr = "127.0.0.1:9000"

func main() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	ctx = context.WithValue(ctx, "key", "client")
	defer cancelFunc()
	conn, err := kgrpc.DialInsecure(ctx,
		kgrpc.WithEndpoint(addr),
		kgrpc.WithMiddleware(
			recovery.Recovery(),
			tracing.Client(),
		),
		kgrpc.WithTimeout(2*time.Second),
		// for tracing remote ip recording
		kgrpc.WithOptions(grpc.WithStatsHandler(&tracing.ClientHandler{})),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	//dial, err := grpc.Dial(addr, grpc.WithInsecure())
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}

	client := version.NewMagusVersionClient(conn)

	reply, err := client.GetVersion(ctx, &version.VersionReq{})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("版本信息: ", reply)
}
