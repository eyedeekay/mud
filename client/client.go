package client

import (
	"context"
	"fmt"
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

func (c *Client) PingPong() error {
	host, err := os.Hostname()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r, err := c.Ping(ctx, &pb.PingRequest{Name: host})
	if err != nil {
		return nil
	}

	c.logger.Info(
		"api response succeed",
		"name", r.Name,
		"token", r.Token,
	)

	return nil
}
