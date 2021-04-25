package processor

import (
	"time"
)

type Config struct {
	// histogram name
	Name string
	// interval between bucket borders
	// everything with timestamp in specific interval aggregates as a slice of histogram
	SliceDuration time.Duration
}

func NewConfig() Config {
	s := Config{}
	s.SliceDuration = time.Minute * 10
	return s
}

func (s Config) WithName(name string) Config {
	s.Name = name
	return s
}

func (s Config) WithSliceDuration(duration time.Duration) Config {
	s.SliceDuration = duration
	return s
}
