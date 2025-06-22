package messagescheduler

import (
	"time"
)

type Scheduler struct {
	ticker *time.Ticker
	c      chan struct{}
	quit   chan struct{}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		c:    make(chan struct{}),
		quit: make(chan struct{})}
}

func (s *Scheduler) Start() <-chan struct{} {
	if s.ticker != nil {
		return s.c
	}

	s.c = make(chan struct{})
	s.quit = make(chan struct{})

	t := time.NewTicker(2 * time.Minute)
	s.ticker = t

	go func() {
		for {
			select {
			case <-t.C:
				select {
				case s.c <- struct{}{}:
				default:
				}
			case <-s.quit:
				t.Stop()
				close(s.c)
				return
			}
		}
	}()

	return s.c
}

func (s *Scheduler) Stop() {
	if s.ticker == nil {
		return
	}
	close(s.quit)
	s.ticker = nil
}
