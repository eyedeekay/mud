package main

import (
	"log"
	"os"

	"github.com/zrma/mud/logging"
	"github.com/zrma/mud/server"
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

	s := server.New(logger, 5555)
	s.Run()
}
