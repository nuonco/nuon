package metrics

import (
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
)

type Incr struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func (m Incr) Write(mw Writer) {
	mw.Incr(m.Name, m.Tags)
}

type Decr struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func (m Decr) Write(mw Writer) {
	mw.Decr(m.Name, m.Tags)
}

type Gauge struct {
	Name  string   `json:"name"`
	Value float64  `json:"value"`
	Tags  []string `json:"tags"`
}

func (m Gauge) Write(mw Writer) {
	mw.Gauge(m.Name, m.Value, m.Tags)
}

type Timing struct {
	Name  string        `json:"name"`
	Value time.Duration `json:"value" swaggertype:"primitive,integer"`
	Tags  []string      `json:"tags"`
}

func (m Timing) Write(mw Writer) {
	mw.Timing(m.Name, m.Value, m.Tags)
}

type Event struct {
	Event *statsd.Event `json:"event"`
}

func (m Event) Write(mw Writer) {
	mw.Event(m.Event)
}
