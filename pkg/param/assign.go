package param

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

// Assign 将src赋值给dest
func Assign(dest, src any) error {
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
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Convert src to int64
		intValue, err := toInt64(src)
		if err != nil {
			return err
		}
		destElem.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Convert src to uint64
		intValue, err := toUint64(src)
		if err != nil {
			return err
		}
		destElem.SetUint(uint64(intValue))
	case reflect.Float32, reflect.Float64:
		// Convert src to float64
		floatValue, err := toFloat64(src)
		if err != nil {
			return err
		}
		destElem.SetFloat(floatValue)
	case reflect.Bool:
		// Convert src to bool
		boolValue, err := toBool(src)
		if err != nil {
			return err
		}
		destElem.SetBool(boolValue)
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

// Convert src to an int64
func toInt64(src any) (int64, error) {
	switch v := src.(type) {
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int(), nil
	case uint, uint8, uint16, uint32, uint64:
		u := reflect.ValueOf(v).Uint()
		if u > math.MaxInt64 {
			return 0, fmt.Errorf("value %d overflows int64", u)
		}
		return int64(u), nil
	case float32, float64:
		return int64(reflect.ValueOf(v).Float()), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}

// Convert src to a uint64
func toUint64(src any) (uint64, error) {
	switch v := src.(type) {
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v).Int()
		if i < 0 {
			return 0, fmt.Errorf("negative value %d cannot be converted to uint64", i)
		}
		return uint64(i), nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint(), nil
	case float32, float64:
		f := reflect.ValueOf(v).Float()
		if f < 0 {
			return 0, fmt.Errorf("negative value %g cannot be converted to uint64", f)
		}
		return uint64(f), nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to uint64", v)
	}
}

func toBool(src any) (bool, error) {
	switch v := src.(type) {
	case bool:
		return v, nil
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return false, fmt.Errorf("cannot convert string %s to bool: %w", err)
		}
		return b, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}
