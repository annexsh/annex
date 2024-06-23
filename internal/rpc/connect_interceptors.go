package rpc

import (
	"connectrpc.com/connect"
)

func WithConnectInterceptors(logger Logger) connect.Option {
	return connect.WithInterceptors(
		//connect.WithRecover() TODO,
		NewConnectLogInterceptor(logger),
	)
}
