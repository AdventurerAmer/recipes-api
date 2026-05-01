package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/joho/godotenv"
)

var Env Environment

const TagName = "cfg"
const EnvVariableSeparator = "_"

func Load(i any) error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("'godotenv.Load' failed: %w", err)
	}
	if err := loadFromEnv(i); err != nil {
		return fmt.Errorf("'loadFromEnv' failed: %w", err)
	}
	return nil
}

func loadFromEnv(i any) error {
	structPtrVal := reflect.ValueOf(i)
	if structPtrVal.Kind() != reflect.Pointer {
		return fmt.Errorf("expected a pointer type got %+v", structPtrVal.Kind())
	}
	structVal := structPtrVal.Elem()
	if structVal.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct type got %+v", structVal.Kind())
	}
	return setFieldsFromEnv(structVal, "")
}

func setFieldsFromEnv(structVal reflect.Value, prefix string) error {
	t := structVal.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		fieldVal := structVal.Field(i)
		if !field.IsExported() {
			continue
		}
		rawTag := field.Tag.Get(TagName)
		if rawTag == "" {
			rawTag = field.Name
		}
		tag := prefix + camelCaseToEnvFmt(rawTag)
		if fieldVal.Kind() == reflect.Struct {
			setFieldsFromEnv(fieldVal, tag+EnvVariableSeparator)
			continue
		}
		envVarVal, ok := os.LookupEnv(tag)
		if !ok {
			return fmt.Errorf("env var %q is not set", tag)
		}
		if err := setFieldValue(fieldVal, envVarVal); err != nil {
			return fmt.Errorf("'setValue' failed: %w", err)
		}
	}
	return nil
}

func setFieldValue(field reflect.Value, value string) error {
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

func camelCaseToEnvFmt(s string) string {
	var (
		parts                []string
		lastUppercaseRuneIdx int
	)

	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			parts = append(parts, s[lastUppercaseRuneIdx:i])
			lastUppercaseRuneIdx = i
		}
	}
	parts = append(parts, s[lastUppercaseRuneIdx:])
	return strings.ToUpper(strings.Join(parts, EnvVariableSeparator))
}
