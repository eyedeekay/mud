package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/zrma/mud/command"
	"github.com/zrma/mud/server"
)

func echo(args ...string) {
	const whitespace = " "
	fmt.Println("당신:", strings.Join(args, whitespace))
}

func main() {
	const (
		linefeed    = '\n'
		linefeedStr = string(linefeed)
		whitespace  = " "
	)

	go func() {
		for {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("] ")
			input, err := reader.ReadString(linefeed)
			if err != nil {
				log.Println(err)
			}

			input = strings.TrimRight(input, linefeedStr)
			inputs := strings.Split(input, whitespace)

			args, token := inputs[:len(inputs)-1], inputs[len(inputs)-1]
			cmd, ok := command.Find(token)
			if !ok {
				fmt.Println("그런 명령어는 찾을 수 없습니다:", input)
				continue
			}

			v, err := cmd.Func()
			if err != nil {
				fmt.Fprintln(os.Stderr, "명령어를 실행하는 도중 에러가 발생했습니다.:", err)
			}

			switch v {
			case command.Exit:
				fmt.Println("접속을 종료합니다.")
				return
			case command.Echo:
				echo(args...)
			}
			fmt.Println(input)
		}
	}()

	s := server.New(5555)
	s.Run()
}
