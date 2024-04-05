package envReader

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

func GetEnvVariable(structEnv interface{}) error {
	val := reflect.ValueOf(structEnv)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return errors.New("input must be a pointer to a struct")
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		structField := typ.Field(i)
		envTag := structField.Tag.Get("env")

		if envTag == "" {
			return fmt.Errorf("tag env must be every fields")
		}

		envVal := os.Getenv(envTag)
		if envVal == "" {
			return fmt.Errorf("variable %v not found", envTag)
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(envVal)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal, err := strconv.ParseInt(envVal, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to convert env var '%s' value '%s' to int: %w", envTag, envVal, err)
			}
			field.SetInt(intVal)
		}

	}

	return nil
}
