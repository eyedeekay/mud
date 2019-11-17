package session

import "sync"

func New() *Session {
	return &Session{}
}

type Session struct {
	sync.Mutex

	msg []string
}

func (s *Session) Put(msg string) {
	s.Lock()
	defer s.Unlock()

	s.msg = append(s.msg, msg)
}

func (s *Session) Get() []string {
	s.Lock()
	defer s.Unlock()

	if len(s.msg) == 0 {
		return nil
	}

	msg := s.msg[:]
	s.msg = s.msg[:0]

	return msg
}
