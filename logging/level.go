package logging

import (
	"encoding/json"
	"fmt"
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

var LevelValues = []Level{LevelDebug, LevelInfo, LevelWarn, LevelError}

func (l Level) String() string {
	return string(l)
}

func (l Level) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

func (l *Level) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	level, err := ParseLevel(s)
	if err != nil {
		return fmt.Errorf("'ParseLevel' failed: %w", err)
	}
	*l = level

	return nil
}

func ParseLevel(s string) (Level, error) {
	switch s {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	default:
		return "", fmt.Errorf("invalid level: %s", s)
	}
}
