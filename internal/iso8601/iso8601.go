package iso8601

import (
	"fmt"
	"time"
)

type ISO8601 struct {
	time.Time
}

func (this *ISO8601) UnmarshalJSON(bytes []byte) error {
	value := string(bytes)
	value = value[1 : len(value)-1] // Trim quotes

	_time, err := time.Parse("2006-01-02T15:04Z", value)
	if err != nil {
		return fmt.Errorf("Error parsing ISO-8601 %q: %w", value, err)
	}

	this.Time = _time

	return nil
}
