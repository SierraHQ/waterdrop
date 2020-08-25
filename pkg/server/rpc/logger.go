package rpc

import (
	"context"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"google.golang.org/grpc/peer"

	"google.golang.org/grpc"
)

func (s *Server) logger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		now := time.Now()
		var ip string
		if peer, ok := peer.FromContext(ctx); ok {
			ip = peer.Addr.String()
		}

		var quota float64
		if deadline, ok := ctx.Deadline(); ok {
			quota = time.Until(deadline).Seconds()
		}

		// call server interceptor
		resp, err = handler(ctx, req)
		var errmsg, retcode string
		if err != nil {
			estatus := status.ExtractStatus(err)
			retcode = estatus.Error()
			errmsg = estatus.Message()
		}

		duration := time.Since(now)

		fields := make([]log.Field, 0, 7)
		fields = append(
			fields,
			log.String("peer_ip", ip),
			log.String("method", info.FullMethod),
			log.Any("req", req),
			log.Float64("quota", quota),
			log.Float64("duration", duration.Seconds()),
			log.String("code", retcode),
			log.String("error", errmsg),
		)

		if duration >= s.config.SlowRequestDuration {
			log.Warn(ctx, "grpc-slow-access-log", fields...)
		} else {
			log.Info(ctx, "grpc-access-log", fields...)
		}

		return
	}
}

func (c *Client) logger() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		now := time.Now()

		var peerInfo peer.Peer
		opts = append(opts, grpc.Peer(&peerInfo))

		var quota float64
		if deadline, ok := ctx.Deadline(); ok {
			quota = time.Until(deadline).Seconds()
		}

		// call client interceptor
		err = invoker(ctx, method, req, reply, cc, opts...)
		var errmsg, retcode string
		if err != nil {
			estatus := status.ExtractStatus(err)
			retcode = estatus.Error()
			errmsg = estatus.Message()
		}

		duration := time.Since(now)

		fields := make([]log.Field, 0, 7)
		fields = append(
			fields,
			log.String("peer_ip", peerInfo.Addr.String()),
			log.String("method", method),
			log.Any("req", req),
			log.Float64("quota", quota),
			log.Float64("duration", duration.Seconds()),
			log.String("code", retcode),
			log.String("error", errmsg),
		)

		if duration >= c.config.SlowRequestDuration {
			log.Warn(ctx, "grpc-slow-access-log", fields...)
		} else {
			log.Info(ctx, "grpc-access-log", fields...)
		}

		return
	}
}
