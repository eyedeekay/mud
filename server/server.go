package server

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pborman/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/zrma/mud/logging"
	"github.com/zrma/mud/pb"
	"github.com/zrma/mud/server/session"
)

func New(logger logging.Logger, port int) *Server {
	s := Server{
		logger:  logger,
		port:    port,
		session: make(map[string]*session.Session),
	}
	return &s
}

type Server struct {
	logger logging.Logger
	port   int

	server *grpc.Server

	mutex   sync.Mutex
	session map[string]*session.Session
	msg     []string
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

const (
	key = "mud"
)

func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingReply, error) {
	s.logger.Info(
		"receive",
		"method", "Ping",
		"name", req.GetName(),
	)

	name := req.GetName()
	token := req.GetToken()
	if token == "" {
		func() {
			s.mutex.Lock()
			defer s.mutex.Unlock()

			token = uuid.New()
			s.session[token] = session.New()
		}()
	}

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name":  name,
		"token": token,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := newToken.SignedString([]byte(key))
	if err != nil {
		return nil, err
	}
	res := &pb.PingReply{
		Name:  name,
		Token: tokenString,
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

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	newToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(key), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := newToken.Claims.(jwt.MapClaims); ok && newToken.Valid {
		s.logger.Info(
			"decrypted",
			"method", "Message",
			"name", claims["name"],
			"token", claims["token"],
		)
	} else {
		return nil, err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.msg = append(s.msg, msg)

	return &pb.MessageReply{}, nil
}

func (s *Server) Receive(req *pb.ReceiveRequest, stream pb.Mud_ReceiveServer) error {
	ticker := time.NewTicker(time.Millisecond * 300)
	defer ticker.Stop()

	for stream.Context().Err() == nil {
		select {
		case <-ticker.C:
			if err := func() error {
				s.mutex.Lock()
				defer s.mutex.Unlock()

				for _, msg := range s.msg {
					if err := stream.Send(&pb.ReceiveReply{
						Msg: msg,
					}); err != nil {
						return err
					}
				}
				s.msg = s.msg[:0]
				return nil
			}(); err != nil {
				return err
			}
		}
	}
	return nil
}
