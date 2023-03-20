package available

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

type TokenBucketLimiter struct {
	tokens chan struct{}
	closeC chan struct{}
}

func NewTokenBucketLimiter(capacity int, interval time.Duration) *TokenBucketLimiter {
	tokens := make(chan struct{}, capacity)
	closeC := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)

		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				select {
				case tokens <- struct{}{}:
				default:
				}
			case <-closeC:
				return
			}
		}

	}()

	return &TokenBucketLimiter{
		tokens: tokens,
		closeC: closeC,
	}

}

func (t TokenBucketLimiter) BuildUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-t.tokens:
			return handler(ctx, req)
		case <-t.closeC:
			// 相当于把限流关了，直接处理请求
			return handler(ctx, req)
		}

	}
}

func (t TokenBucketLimiter) Close() error {
	close(t.closeC)
	return nil
}
