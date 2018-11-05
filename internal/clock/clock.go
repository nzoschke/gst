package clock

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// Clock is common time methods
type Clock interface {
	Now() time.Time
	TimestampNow() *timestamp.Timestamp
}

var _ Clock = (*RealClock)(nil) // assert interface

// RealClock is a real time implementation
type RealClock struct{}

// Now returns a time for now
func (RealClock) Now() time.Time { return time.Now() }

// TimestampNow returns a proto timestamp for now
func (RealClock) TimestampNow() *timestamp.Timestamp { return ptypes.TimestampNow() }

// TestClock is a mock time implementation
type TestClock struct{}

// Now returns a time for 0
func (c TestClock) Now() time.Time { return time.Unix(0, 0) }

// TimestampNow returns a proto timestamp for 0
func (c TestClock) TimestampNow() *timestamp.Timestamp {
	t, _ := ptypes.TimestampProto(time.Unix(0, 0))
	return t
}
