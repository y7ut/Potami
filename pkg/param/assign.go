package param

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// assign copies the value of src into dest.
func assign(dest, src any) error {
	// Check if dest is a pointer
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr || destValue.IsNil() {
		return errors.New("dest must be a non-nil pointer")
	}

	// Get the element of dest (actual value it points to)
	destElem := destValue.Elem()

	// Handle different dest types
	switch destElem.Kind() {
	case reflect.String:
		// Convert src to string
		strValue, err := toString(src)
		if err != nil {
			return err
		}
		destElem.SetString(strValue)
	case reflect.Float64:
		// Convert src to float64
		floatValue, err := toFloat64(src)
		if err != nil {
			return err
		}
		destElem.SetFloat(floatValue)
	default:
		return fmt.Errorf("unsupported dest type: %s", destElem.Kind())
	}

	return nil
}

// toString converts src to a string
func toString(src any) (string, error) {
	switch v := src.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case fmt.Stringer:
		return v.String(), nil
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v), nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%f", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	default:
		return "", fmt.Errorf("cannot convert %T to string", v)
	}
}

// toFloat64 converts src to a float64
func toFloat64(src any) (float64, error) {
	switch v := src.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(v).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(v).Uint()), nil
	case string:
		// Attempt to parse string as float64
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}
