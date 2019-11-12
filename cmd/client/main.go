package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/zrma/mud/client"
	"github.com/zrma/mud/command"
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

	const (
		lf         = '\n'
		cr         = '\r'
		lfStr      = string(lf)
		crStr      = string(cr)
		whitespace = " "
	)

	authToken, err := c.PingPong()
	if err != nil {
		logger.Err(
			"api request failed",
			"method", "Ping",
			"err", err,
		)
		return
	}

	var mutex sync.RWMutex

	waitCh := make(chan interface{})
	defer close(waitCh)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer func() {
			defer cancel()
			waitCh <- nil
		}()

		token := authToken

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for ctx.Err() == nil {
			select {
			case <-ticker.C:
				t, err := c.PingPong()
				if err != nil {
					logger.Err(
						"api request failed",
						"method", "Ping",
						"err", err,
					)
				}
				if token != t {
					token = t
					func(t string) {
						mutex.Lock()
						defer mutex.Unlock()
						authToken = t
					}(t)
				}
			case <-ctx.Done():
				break
			}
		}
	}()

	for ctx.Err() == nil {
		reader := bufio.NewReader(os.Stdin)

		input, err := reader.ReadString(lf)
		if err != nil {
			if err == io.EOF {
				logger.Info(
					"input continue",
					"err", err,
				)
				continue
			}
			logger.Err(
				"input failed",
				"err", err,
			)
			return
		}

		input = strings.TrimRight(input, lfStr)
		input = strings.TrimRight(input, crStr)
		inputs := strings.Split(input, whitespace)

		_, token := inputs[:len(inputs)-1], inputs[len(inputs)-1]
		cmd, ok := command.Find(token)
		if !ok {
			fmt.Println("그런 명령어는 찾을 수 없습니다:", input)
			continue
		}

		v, err := cmd.Func()
		if err != nil {
			fmt.Println("명령어를 실행하는 도중 에러가 발생했습니다.:", err)
		}

		switch v {
		case command.Exit:
			fmt.Println("접속을 종료합니다.")
			cancel()
			break
		case command.Echo:
			func(msg string) {
				mutex.RLock()
				defer mutex.RUnlock()
				if err := c.SendMessage(authToken, msg); err != nil {
					logger.Err(
						"api request failed",
						"method", "Message",
						"err", err,
					)
				}
			}(input)
		}
	}

	<-waitCh

	time.Sleep(time.Second)
	logger.Info("end")
}
