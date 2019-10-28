package command

import (
	"errors"
	"fmt"
)

type OpCode int

const (
	Exit OpCode = iota
	Echo
)

type command struct {
	Word string
	Func func() (OpCode, error)
}

var commands map[string]*command

func Register(word string, f func() (OpCode, error)) error {
	if commands == nil {
		commands = make(map[string]*command)
	}

	if _, ok := commands[word]; ok {
		return errors.New(fmt.Sprintln("already registered command", word))
	}

	commands[word] = &command{
		Word: word,
		Func: f,
	}
	return nil
}

func Find(word string) (*command, bool) {
	cmd, ok := commands[word]
	return cmd, ok
}

var _ = Register("끝", func() (o OpCode, e error) {
	return Exit, nil
})

var _ = Register("말", func() (o OpCode, e error) {
	return Echo, nil
})
