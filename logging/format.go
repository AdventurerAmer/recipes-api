package logging

import (
	"encoding/json"
	"fmt"
)

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

var FormatValues = []Format{FormatText, FormatJSON}

func (f Format) String() string {
	return string(f)
}

func (f Format) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func (f *Format) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "text":
		*f = FormatText
	case "json":
		*f = FormatJSON
	default:
		return fmt.Errorf("invalid format: %s", s)
	}

	return nil
}

func ParseFormat(s string) (Format, error) {
	switch s {
	case "text":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("invalid format: %s", s)
	}
}
