package server

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/pborman/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/zrma/mud/logging"
	"github.com/zrma/mud/pb"
)

func New(logger logging.Logger, port int) *Server {
	s := Server{logger: logger, port: port}
	return &s
}

type Server struct {
	logger logging.Logger
	port   int

	server *grpc.Server
}

func (s *Server) Run() {
	server, err := net.Listen("tcp", ":"+strconv.Itoa(s.port))
	if err != nil {
		panic("couldn't start listening: " + err.Error())
	}

	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     3 * time.Minute,
			Time:                  1 * time.Minute,
			Timeout:               10 * time.Second,
			MaxConnectionAge:      5 * time.Minute,
			MaxConnectionAgeGrace: 30 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             20 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	s.server = grpc.NewServer(opts...)

	pb.RegisterMudServer(s.server, s)
	if err := s.server.Serve(server); err != nil {
		panic("failed to serve: " + err.Error())
	}
}

func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingReply, error) {
	s.logger.Info(
		"receive",
		"method", "Ping",
		"name", req.GetName(),
	)

	res := &pb.PingReply{
		Name:  req.GetName(),
		Token: uuid.New(),
	}

	return res, nil
}

func (s *Server) Message(ctx context.Context, req *pb.MessageRequest) (*pb.MessageReply, error) {
	token := req.GetToken()
	msg := req.GetMsg()

	s.logger.Info(
		"receive",
		"method", "Message",
		"token", token,
		"msg", msg,
	)

	return &pb.MessageReply{}, nil
}
