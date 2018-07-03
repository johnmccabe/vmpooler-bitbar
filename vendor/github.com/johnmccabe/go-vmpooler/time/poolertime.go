package time

import (
	"fmt"
	"strings"
	"time"
)

const poolerTimeLayout = "2006-01-02 15:04:05 -0700"

// PoolerTime is the time format returned from vmpooler
type PoolerTime struct {
	time.Time
}

// UnmarshalJSON provides custom unmarshalling since vmpooler time isn't RFC 3339 compliant
func (pt *PoolerTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		pt.Time = time.Time{}
		return
	}
	pt.Time, err = time.Parse(poolerTimeLayout, s)
	return
}

var nilTime = (time.Time{}).UnixNano()

// MarshalJSON provides custom marshalling since vmpooler time isn't RFC 3339 compliant
func (pt *PoolerTime) MarshalJSON() ([]byte, error) {
	if pt.Time.UnixNano() == nilTime {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", pt.Time.Format(poolerTimeLayout))), nil
}

// IsSet checks if the time is not the nil time
func (pt *PoolerTime) IsSet() bool {
	return pt.UnixNano() != nilTime
}
