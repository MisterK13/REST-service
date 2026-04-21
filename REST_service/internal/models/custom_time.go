package models

import (
	"fmt"
	"strings"
	"time"
)

type CustomTime struct {
	time.Time
}

const layout = "01-2006"

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		ct.Time = time.Time{}
		return nil
	}

	t, err := time.Parse(layout, s)
	if err != nil {
		return fmt.Errorf("invalid date format, expected MM-YYYY: %w", err)
	}
	ct.Time = t
	return nil
}

func (ct CustomTime) MarshalJSON() ([]byte, error) {
	if ct.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", ct.Time.Format(layout))), nil
}

func (ct *CustomTime) UnmarshalParam(param string) error {
	if param == "" {
		ct.Time = time.Time{}
		return nil
	}

	t, err := time.Parse(layout, param)
	if err != nil {
		return fmt.Errorf("invalid date format, expected MM-YYYY: %w", err)
	}
	ct.Time = t
	return nil
}
