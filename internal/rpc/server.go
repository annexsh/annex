package rpc

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	connectcors "connectrpc.com/cors"
	"connectrpc.com/grpcreflect"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	connectPath = "/connect"
	grpcPath    = "/" // grpc does not support a custom path
)

type Server struct {
	addr            string
	mux             *http.ServeMux
	grpcSrv         *grpc.Server
	grpcOptions     []grpc.ServerOption
	httpSrv         *http.Server
	grpcServices    []grpcSvcRegistrar
	connectSvcNames []string
}

func NewServer(address string) *Server {
	mux := http.NewServeMux()
	return &Server{
		addr: address,
		mux:  mux,
	}
}

func (s *Server) RegisterConnect(path string, handler http.Handler, corsOrigins ...string) {
	if len(corsOrigins) > 0 {
		c := cors.New(cors.Options{
			AllowedOrigins: corsOrigins,
			AllowedMethods: connectcors.AllowedMethods(),
			AllowedHeaders: connectcors.AllowedHeaders(),
			ExposedHeaders: connectcors.ExposedHeaders(),
		})
		handler = c.Handler(handler)
	}

	s.mux.Handle(connectPath+path, http.StripPrefix(connectPath, handler))
	svcName := strings.TrimPrefix("/", path)
	svcName = strings.TrimSuffix(svcName, "/")
	s.connectSvcNames = append(s.connectSvcNames, svcName)
}

type grpcSvcRegistrar struct {
	desc    *grpc.ServiceDesc
	service any
}

func (s *Server) RegisterGRPC(desc *grpc.ServiceDesc, service any) {
	s.grpcServices = append(s.grpcServices, grpcSvcRegistrar{
		desc:    desc,
		service: service,
	})
}

func (s *Server) WithGRPCOptions(opt ...grpc.ServerOption) {
	s.grpcOptions = append(s.grpcOptions, opt...)
}

func (s *Server) Serve() error {
	numConnect := len(s.connectSvcNames)
	numGRPC := len(s.grpcServices)

	if numConnect == 0 && numGRPC == 0 {
		return errors.New("must register at least one service")
	}

	if numGRPC > 0 {
		s.grpcSrv = grpc.NewServer(s.grpcOptions...)
		for _, svc := range s.grpcServices {
			s.grpcSrv.RegisterService(svc.desc, svc.service)
		}
		reflection.Register(s.grpcSrv)
		s.mux.Handle(grpcPath, s.grpcSrv)
	}

	s.httpSrv = &http.Server{
		Addr:    s.addr,
		Handler: h2c.NewHandler(s.mux, &http2.Server{}),
	}

	return s.httpSrv.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := s.httpSrv.Shutdown(ctx)
	if s.grpcSrv != nil {
		s.grpcSrv.Stop()
	}
	return err
}

func (s *Server) ConnectAddress() string {
	return "http://" + s.addr + connectPath
}

func (s *Server) GRPCAddress() string {
	return s.addr
}

func (s *Server) registerConnectReflection() {
	reflector := grpcreflect.NewStaticReflector(s.connectSvcNames...)
	s.mux.Handle(grpcreflect.NewHandlerV1(reflector))
	s.mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
}
