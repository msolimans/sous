package logging

import "time"

var zeroTime = time.Time{}

// A MessageInterval is used where messages refer to an event that spans an interval of time.
type MessageInterval struct {
	start, end time.Time
}

// OpenInterval returns "the beginning" of an interval - on which hasn't completed yet.
func OpenInterval(start time.Time) MessageInterval {
	return MessageInterval{
		start: start,
	}
}

// CompleteInterval returns and interval that closes at the time of call.
func CompleteInterval(start time.Time) MessageInterval {
	return MessageInterval{
		start: start,
		end:   time.Now(),
	}
}

// NewInterval allows callers to completely specify start and end times.
func NewInterval(start, end time.Time) MessageInterval {
	return MessageInterval{
		start: start,
		end:   end,
	}
}

// Complete queries whether the interval is closed - in other words: does it have an end time?
func (i MessageInterval) Complete() bool {
	return i.end != zeroTime
}

// EachField implements eachFielder on MessageInterval.
// In general, messages that include a MessageInterval should pass the
// FieldReportFn they receive on their own EachField to their interval.
func (i MessageInterval) EachField(fn FieldReportFn) {
	fn("started-at", i.start.UTC().Format(time.RFC3339))
	if i.Complete() {
		fn("finished-at", i.end.UTC().Format(time.RFC3339))
	}
	if i.end.After(i.start) {
		fn("duration", int64(i.end.Sub(i.start)/time.Microsecond))
	} else {
		fn("duration", int64(0))
	}
}
