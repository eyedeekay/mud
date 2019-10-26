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

var _ = Register("끝", func() (op, error) {
	return Exit, nil
})

func main() {
	const (
		delimiterRune = '\n'
		delimiterStr  = string(delimiterRune)
	)

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("] ")
		text, err := reader.ReadString(delimiterRune)
		if err != nil {
			log.Println(err)
		}

		text = strings.TrimRight(text, delimiterStr)
		cmd, ok := commands[text]
		if !ok {
			fmt.Println("그런 명령어는 찾을 수 없습니다:", text)
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
		}
		fmt.Println(text)
	}
}
