package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/zrma/mud/logging"
	"github.com/zrma/mud/pb"
)

func New(logger logging.Logger, host string, port int) *Client {
	c := Client{logger: logger, host: host, port: port}
	return &c
}

type Client struct {
	logger logging.Logger
	host   string
	port   int

	conn *grpc.ClientConn
	pb.MudClient
}

func (c *Client) Init() error {
	address := fmt.Sprintf("%s:%s", c.host, strconv.Itoa(c.port))

	// Set up a connection to the server.
	conn, err := grpc.Dial(
		address,
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			// keepalive settings - https://github.com/grpc/grpc/blob/master/doc/keepalive.md
			Time:                1 * time.Minute,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return err
	}

	c.conn = conn
	c.MudClient = pb.NewMudClient(conn)
	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) PingPong() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := c.Ping(ctx, &pb.PingRequest{Name: host})
	if err != nil {
		return "", err
	}

	c.logger.Info(
		"api response succeed",
		"name", r.GetName(),
		"token", r.GetToken(),
	)

	return r.GetToken(), nil
}

func (c *Client) SendMessage(token, msg string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := c.Message(ctx, &pb.MessageRequest{
		Token: token,
		Msg:   msg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Subscribe(ctx context.Context, token string, f func(string) error) error {
	stream, err := c.Receive(ctx, &pb.ReceiveRequest{
		Token: token,
	})
	if err != nil {
		return err
	}

	for ctx.Err() == nil {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := f(r.GetMsg()); err != nil {
			return err
		}
	}

	return nil
}
