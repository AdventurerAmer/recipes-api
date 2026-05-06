package config

import (
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/joho/godotenv"
)

type Option = func(cfg *Config) error

func WithTagName(tagName string) Option {
	return func(cfg *Config) error {
		if tagName == "" {
			return fmt.Errorf("'tagName' is empty")
		}
		cfg.TagName = tagName
		return nil
	}
}

func WithSepartor(separator string) Option {
	return func(cfg *Config) error {
		if separator == "" {
			return fmt.Errorf("'separator' is empty")
		}
		cfg.Separator = separator
		return nil
	}
}

func WithPrefix(prefix string) Option {
	return func(cfg *Config) error {
		cfg.Prefix = strings.ToUpper(prefix)
		return nil
	}
}

type Config struct {
	TagName   string
	Separator string
	Prefix    string
}

func Load(v any, opts ...Option) error {
	cfg := &Config{
		TagName:   "cfg",
		Separator: "_",
		Prefix:    "",
	}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return fmt.Errorf("set option failed: %w", err)
		}
	}
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("'godotenv.Load' failed: %w", err)
	}
	if err := loadFromEnv(v, cfg); err != nil {
		return fmt.Errorf("'loadFromEnv' failed: %w", err)
	}
	return nil
}

func loadFromEnv(v any, cfg *Config) error {
	structPtrVal := reflect.ValueOf(v)
	if structPtrVal.Kind() != reflect.Pointer {
		return fmt.Errorf("expected a pointer type got %+v", structPtrVal.Kind())
	}
	structVal := structPtrVal.Elem()
	if structVal.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct type got %+v", structVal.Kind())
	}
	prefix := cfg.Prefix
	if prefix != "" && !strings.HasSuffix(cfg.Prefix, cfg.Separator) {
		prefix += cfg.Separator
	}
	return setFieldsFromEnv(structVal, cfg, prefix)
}

func setFieldsFromEnv(structVal reflect.Value, cfg *Config, prefix string) error {
	t := structVal.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		fieldVal := structVal.Field(i)
		if !field.IsExported() {
			continue
		}
		rawTag := field.Tag.Get(cfg.TagName)
		if rawTag == "" {
			rawTag = field.Name
		}
		tag := prefix + camelCaseToEnvFmt(rawTag, cfg)
		if fieldVal.Kind() == reflect.Struct {
			setFieldsFromEnv(fieldVal, cfg, tag+cfg.Separator)
			continue
		}
		varVal, ok := os.LookupEnv(tag)
		if !ok {
			return fmt.Errorf("env-var: %q is not set", tag)
		}
		if err := setFieldValue(fieldVal, varVal); err != nil {
			return fmt.Errorf("'setFieldValue' failed: %w", err)
		}
	}
	return nil
}

func setFieldValue(field reflect.Value, value string) error {
	switch field.Interface().(type) {
	case Environment:
		if !slices.Contains(Environments, Environment(value)) {
			return fmt.Errorf("supported environment expected one of %+v", Environments)
		}
		field.Set(reflect.ValueOf(Environment(value)))
		return nil
	case time.Duration:
		d, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("'time.ParseDuration' failed: %w", err)
		}
		field.Set(reflect.ValueOf(d))
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("'strconv.ParseInt' failed: %w", err)
		}
		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("'strconv.ParseUint' failed: %w", err)
		}
		field.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("'strconv.ParseFloat' failed: %w", err)
		}
		field.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("'strconv.ParseBool' failed: %w", err)
		}
		field.SetBool(b)
	default:
		return fmt.Errorf("unsupported type: %s", field.Kind())
	}

	return nil
}

func camelCaseToEnvFmt(s string, cfg *Config) string {
	var (
		parts                []string
		lastUppercaseRuneIdx int
	)
	var prev rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) && unicode.IsLower(prev) {
			parts = append(parts, s[lastUppercaseRuneIdx:i])
			lastUppercaseRuneIdx = i
		}
		prev = r
	}
	parts = append(parts, s[lastUppercaseRuneIdx:])
	return strings.ToUpper(strings.Join(parts, cfg.Separator))
}
