package clock

import (
	"fmt"
	"time"

	"github.com/beevik/ntp"
	"github.com/gotd/td/clock"
)

const defaultHost = "pool.ntp.org"

type networkClock struct {
	offset time.Duration
}

func (n *networkClock) Now() time.Time {
	return time.Now().Add(n.offset)
}

func (n *networkClock) Timer(d time.Duration) clock.Timer {
	return clock.System.Timer(d)
}

func (n *networkClock) Ticker(d time.Duration) clock.Ticker {
	return clock.System.Ticker(d)
}

// New default ntp host is 'pool.ntp.org'
func New(ntpHost ...string) (clock.Clock, error) {
	var host string
	switch len(ntpHost) {
	case 0:
		host = defaultHost
	case 1:
		host = ntpHost[0]
	default:
		return nil, fmt.Errorf("too many ntp hosts")
	}

	resp, err := ntp.Query(host)
	if err != nil {
		return nil, err
	}

	return &networkClock{
		offset: resp.ClockOffset,
	}, nil
}
