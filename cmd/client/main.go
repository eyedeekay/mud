package main

import (
	"log"
	"os"
	"time"

	"github.com/zrma/mud/client"
	"github.com/zrma/mud/logging"
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

	c := client.New(logger, host, port)
	if err := c.Init(); err != nil {
		logger.Err(
			"client initializing failed",
			"err", err,
		)
		return
	}
	defer func() {
		if err := c.Close(); err != nil {
			logger.Err(
				"client closing failed",
				"err", err,
			)
		}
	}()

	if err := c.PingPong(); err != nil {
		logger.Err(
			"api request failed",
			"method", "Ping",
			"err", err,
		)
	}

	time.Sleep(time.Second)
	logger.Info("end")
}
