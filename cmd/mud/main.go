package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type op int

const (
	Exit op = iota
	Echo
)

type command struct {
	word string
	f    func() (op, error)
}

var commands map[string]command

func Register(word string, f func() (op, error)) error {
	if commands == nil {
		commands = make(map[string]command)
	}

	if _, ok := commands[word]; ok {
		return errors.New(fmt.Sprintln("already registered command", word))
	}

	commands[word] = command{
		word: word,
		f:    f,
	}
	return nil
}

var _ = Register("끝", func() (o op, e error) {
	return Exit, nil
})

var _ = Register("말", func() (o op, e error) {
	return Echo, nil
})

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
		cmd, ok := commands[token]
		if !ok {
			fmt.Println("그런 명령어는 찾을 수 없습니다:", input)
			continue
		}

		v, err := cmd.f()
		if err != nil {
			fmt.Fprintln(os.Stderr, "명령어를 실행하는 도중 에러가 발생했습니다.:", err)
		}

		switch v {
		case Exit:
			fmt.Println("접속을 종료합니다.")
			return
		case Echo:
			echo(args...)
		}
		fmt.Println(input)
	}
}
