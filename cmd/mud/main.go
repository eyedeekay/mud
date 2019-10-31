package main

import (
	"go.uber.org/zap"

	"github.com/zrma/mud/server"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	sugar.Infow(
		"start",
		"method", "main",
	)
	s := server.New(logger, 5555)
	s.Run()
}
