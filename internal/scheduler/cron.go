package scheduler

import (
	"time"

	"github.com/robfig/cron/v3"
)

type CronService struct {
	parser cron.Parser
}

func NewCronService() *CronService {
	return &CronService{
		parser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor),
	}
}

func (s *CronService) Validate(expr string) error {
	_, err := s.parser.Parse(expr)
	return err
}

func (s *CronService) Next(expr string, from time.Time) (time.Time, error) {
	sched, err := s.parser.Parse(expr)
	if err != nil {
		return time.Time{}, err
	}
	return sched.Next(from), nil
}

func (s *CronService) Preview(expr string, from time.Time, count int) ([]time.Time, error) {
	if count <= 0 {
		count = 5
	}
	sched, err := s.parser.Parse(expr)
	if err != nil {
		return nil, err
	}
	out := make([]time.Time, 0, count)
	cursor := from
	for i := 0; i < count; i++ {
		next := sched.Next(cursor)
		out = append(out, next)
		cursor = next
	}
	return out, nil
}
