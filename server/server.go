package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/zrma/mud/command"
)

type Server struct {
	logger *zap.Logger
	port   int
}

func New(logger *zap.Logger, port int) *Server {
	s := Server{logger, port}
	return &s
}

func (s Server) Run() {

	server, err := net.Listen("tcp", ":"+strconv.Itoa(s.port))
	if server == nil {
		panic("couldn't start listening: " + err.Error())
	}

	conn := clientConn(s.logger, server)
	for {
		go handleConn(s.logger, <-conn)
	}
}

func clientConn(logger *zap.Logger, listener net.Listener) chan net.Conn {
	ch := make(chan net.Conn)
	i := 0

	go func() {
		sugar := logger.Sugar()

		for {
			client, err := listener.Accept()
			if client == nil {
				sugar.Errorw(
					"couldn't accept",
					"err", err,
				)
				continue
			}

			i++
			sugar.Infow(
				"connect",
				"count", i,
				"addr", fmt.Sprintf("%v <-> %v\n", client.LocalAddr(), client.RemoteAddr()),
			)

			ch <- client
		}
	}()

	return ch
}

func echo(w io.Writer, args ...string) {
	const whitespace = " "
	fmt.Fprintln(w, "당신:", strings.Join(args, whitespace))
}

func handleConn(logger *zap.Logger, client net.Conn) {
	sugar := logger.Sugar()
	defer func() {
		if err := client.Close(); err != nil {
			sugar.Errorw(
				"client close failed",
				"method", "close",
				"err", err,
			)
		}
		sugar.Infow(
			fmt.Sprintf("disconnected: %v\n", client.RemoteAddr()),
			"method", "close",
		)
	}()

	const (
		lf         = '\n'
		cr         = '\r'
		lfStr      = string(lf)
		crStr      = string(cr)
		whitespace = " "
	)

	for {
		reader := bufio.NewReader(client)
		fmt.Fprintf(client, "] ")
		input, err := reader.ReadString(lf)
		if err != nil {
			if err == io.EOF {
				return
			}
			sugar.Errorw(
				"client read failed",
				"err", err,
			)
			return
		}

		input = strings.TrimRight(input, lfStr)
		input = strings.TrimRight(input, crStr)
		inputs := strings.Split(input, whitespace)

		args, token := inputs[:len(inputs)-1], inputs[len(inputs)-1]
		cmd, ok := command.Find(token)
		if !ok {
			fmt.Fprintln(client, "그런 명령어는 찾을 수 없습니다:", input)
			continue
		}

		v, err := cmd.Func()
		if err != nil {
			fmt.Fprintln(os.Stderr, "명령어를 실행하는 도중 에러가 발생했습니다.:", err)
		}

		switch v {
		case command.Exit:
			fmt.Fprintln(client, "접속을 종료합니다.")
			return
		case command.Echo:
			echo(client, args...)
		}
		fmt.Println(input)
	}
}
