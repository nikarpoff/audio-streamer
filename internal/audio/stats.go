package audio

import (
	"log"
	"time"
)

type Metric struct {
	count int
	total time.Duration
	max   time.Duration
}

func (m *Metric) Add(d time.Duration) {
	m.count++
	m.total += d
	if d > m.max {
		m.max = d
	}
}

func (m *Metric) Report(label string) {
	if m.count == 0 {
		return
	}
	avg := m.total / time.Duration(m.count)
	log.Printf("[Perf] %-20s avg=%.1f ms max=%.1f ms samples=%d",
		label,
		float64(avg.Microseconds())/1000,
		float64(m.max.Microseconds())/1000,
		m.count,
	)
	m.count, m.total, m.max = 0, 0, 0
}

func Run(m *Metric, label string, delay int16) {
	for range time.Tick(time.Duration(delay) * time.Second) {
		m.Report(label)
	}
}
