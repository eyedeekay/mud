package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/zrma/mud/logging"
	"github.com/zrma/mud/pb"
)

const (
	host = "localhost"
	port = 5555
)

func main() {
	const (
		dev  = "development"
		prod = "production"
		skip = "skip"
	)

	var logLevel logging.LogLevel
	switch os.Getenv("environment") {
	case prod:
		logLevel = logging.Prod
	case skip:
		logLevel = logging.None
	case dev:
		fallthrough
	default:
		logLevel = logging.Dev
	}

	logger, err := logging.NewLogger(logLevel)
	if err != nil {
		log.Fatalln(err)
	}
	logger.Info(
		"start",
		"method", "main",
	)

	address := fmt.Sprintf("%s:%s", host, strconv.Itoa(port))

	// Set up a connection to the server.
	conn, err := grpc.Dial(
		address,
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			// keepalive settings - https://github.com/grpc/grpc/blob/master/doc/keepalive.md
			Time:                60 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		logger.Err(
			"connecting failed",
			"err", err,
		)
		return
	}
	defer conn.Close()
	c := pb.NewMudClient(conn)

	host, err := os.Hostname()
	if err != nil {
		logger.Err(
			"getting hostname failed",
			"err", err,
		)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r, err := c.Ping(ctx, &pb.PingRequest{Name: host})
	if err != nil {
		logger.Err(
			"api request failed",
			"method", "Ping",
			"err", err,
		)
	}
	logger.Info(
		"api response succeed",
		"name", r.Name,
		"token", r.Token,
	)

	time.Sleep(time.Second)
	logger.Info("end")
}
