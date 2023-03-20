package available

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

type FixedWindowLimiter struct {
	// ts 当前窗口开始时间
	ts int64
	// internal 窗口大小
	internal int64
	rate     int64
	cnt      int64
}

func NewFixedWindowLimiter(internal time.Duration, rate int64) *FixedWindowLimiter {

	return &FixedWindowLimiter{
		ts:       time.Now().UnixMilli(),
		internal: internal.Milliseconds(),
		rate:     rate,
	}
}

func (f *FixedWindowLimiter) BuildUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		// 先计算是否需要重开窗口，需要的话重置ts和cnt
		now := time.Now().UnixMilli()

		if f.ts+f.internal < now {
			f.ts = now
			f.cnt = 0
		}

		if f.cnt >= f.rate {

			return
		}

		// 判断是否达到阈值

		return handler(ctx, req)
	}
}
