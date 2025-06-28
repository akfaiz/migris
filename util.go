package schema

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strconv"
)

func isEmptyString(s string) bool {
	return len(s) == 0
}

func optionalInt(defaultValue int, values ...int) int {
	if len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

func optionalString(defaultValue string, values ...string) string {
	if len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

func optionalBool(defaultValue bool, values ...bool) bool {
	if len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

func execContext(ctx context.Context, tx *sql.Tx, sqls ...string) error {
	for _, sql := range sqls {
		if debug {
			log.Printf("Executing SQL: %s\n", sql)
		}
		if _, err := tx.ExecContext(ctx, sql); err != nil {
			if debug {
				log.Printf("Error executing SQL: %s\nError: %v\n", sql, err)
			}
			return err
		}
	}
	return nil
}

type arr[T any] []T

func (a *arr[T]) append(values ...T) {
	for _, value := range values {
		*a = append(*a, value)
	}
}

func (a *arr[T]) toSlice() []T {
	return *a
}

func (a *arr[T]) appendIfNotError(values []T, err error) error {
	if err != nil {
		return err
	}
	for _, value := range values {
		*a = append(*a, value)
	}
	return nil
}

func toString(value any) string {
	reflectType := reflect.TypeOf(value)
	if reflectType == nil {
		return ""
	}
	switch reflectType.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			return str
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intValue, ok := value.(int); ok {
			return strconv.Itoa(intValue)
		}
	case reflect.Float32, reflect.Float64:
		if floatValue, ok := value.(float64); ok {
			return strconv.FormatFloat(floatValue, 'f', -1, 64)
		}
	case reflect.Bool:
		if boolValue, ok := value.(bool); ok {
			if boolValue {
				return "true"
			}
			return "false"
		}
	case reflect.Map:
		if mapValue, ok := value.(map[string]any); ok {
			result := "{"
			for k, v := range mapValue {
				result += k + ":" + toString(v) + ","
			}
			if len(result) > 1 {
				result = result[:len(result)-1] // Remove trailing comma
			}
			result += "}"
			return result
		}
	case reflect.Slice, reflect.Array:
		if sliceValue, ok := value.([]any); ok {
			result := "["
			for _, v := range sliceValue {
				result += toString(v) + ","
			}
			if len(result) > 1 {
				result = result[:len(result)-1] // Remove trailing comma
			}
			result += "]"
			return result
		}
	default:
		return fmt.Sprintf("%v", value) // Fallback for other types
	}

	return fmt.Sprintf("%v", value) // Fallback for unsupported types
}
