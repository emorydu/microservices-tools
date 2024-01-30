package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/emorydu/microservices-tools/pkg/logger"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

const (
	maxConnectionIdle = 5
	gRPCTimeout       = 15
	maxConnectionAge  = 15
	gRPCTime          = 10
)

type Config struct {
	Port        string `mapstructure:"port"`
	Host        string `mapstructure:"host"`
	Development bool   `mapstructure:"development"`
}

type Server struct {
	Grpc   *grpc.Server
	Config *Config
	Log    logger.Logger
}

func NewGrpcServer(log logger.Logger, config *Config) *Server {
	unaryServerInterceptors := []grpc.UnaryServerInterceptor{
		otelgrpc.UnaryServerInterceptor(),
	}
	streamServerInterceptors := []grpc.StreamServerInterceptor{
		otelgrpc.StreamServerInterceptor(),
	}

	s := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: maxConnectionIdle * time.Minute,
			Timeout:           gRPCTimeout * time.Second,
			MaxConnectionAge:  maxConnectionAge * time.Minute,
			Time:              gRPCTime * time.Minute,
		}),

		grpc.ChainStreamInterceptor(streamServerInterceptors...),
		grpc.ChainUnaryInterceptor(unaryServerInterceptors...),
	)

	return &Server{Grpc: s, Config: config, Log: log}
}

func (s *Server) RunGrpcServer(ctx context.Context, configGrpc ...func(grpcServer *grpc.Server)) error {
	listen, err := net.Listen("tcp", s.Config.Port)
	if err != nil {
		return errors.Wrap(err, "net.Listen")
	}

	if len(configGrpc) > 0 {
		grpcFunc := configGrpc[0]
		if grpcFunc != nil {
			grpcFunc(s.Grpc)
		}
	}

	if s.Config.Development {
		reflection.Register(s.Grpc)
	}

	if len(configGrpc) > 0 {
		grpcFunc := configGrpc[0]
		if grpcFunc != nil {
			grpcFunc(s.Grpc)
		}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				s.Log.Infof("shutting down grpc PORT: {%s}", s.Config.Port)
				s.shutdown()
				s.Log.Info("grpc exited properly")
				return
			}
		}
	}()

	s.Log.Infof("grpc server is listening on port: %s", s.Config.Port)

	err = s.Grpc.Serve(listen)
	if err != nil {
		s.Log.Error(fmt.Sprintf("[grpcServer_RunGrpcServer.Serve] grpc server serve error: %v", err))
	}

	return err
}

func (s *Server) shutdown() {
	s.Grpc.Stop()
	s.Grpc.GracefulStop()
}
