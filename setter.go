package config

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrUnsupportedType occurs when a type is being parsed that isn't supported
	ErrUnsupportedType = errors.New("an unsupported type configuration was passed in")
)

type setter func(string) error

func getSetter(f reflect.Value) (setter, error) {
	typ := f.Type()

	switch typ.Kind() {
	case reflect.String:
		return stringSetter(f), nil
	case reflect.Bool:
		return boolSetter(f), nil
	case reflect.Float32, reflect.Float64:
		return floatSetter(f, typ), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intSetter(f, typ), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uintSetter(f, typ), nil
	case reflect.Slice:
		return sliceSetter(f, typ), nil
	default:
		return nil, ErrUnsupportedType
	}
}

func stringSetter(f reflect.Value) setter {
	return func(s string) error {
		f.SetString(s)
		return nil
	}
}

func intSetter(f reflect.Value, typ reflect.Type) setter {
	return func(s string) error {
		var (
			val int64
			err error
		)

		if f.Kind() == reflect.Int64 && typ.PkgPath() == "time" && typ.Name() == "Duration" {
			var d time.Duration
			d, err = time.ParseDuration(s)
			val = int64(d)
		} else {
			val, err = strconv.ParseInt(s, 0, typ.Bits())
		}

		if err != nil {
			return err
		}

		f.SetInt(val)
		return nil
	}
}

func uintSetter(f reflect.Value, typ reflect.Type) setter {
	return func(s string) error {
		val, err := strconv.ParseUint(s, 0, typ.Bits())
		if err != nil {
			return err
		}

		f.SetUint(val)
		return nil
	}
}

func boolSetter(f reflect.Value) setter {
	return func(s string) error {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}

		f.SetBool(b)
		return nil
	}
}

func floatSetter(f reflect.Value, typ reflect.Type) setter {
	return func(s string) error {
		fl, err := strconv.ParseFloat(s, typ.Bits())
		if err != nil {
			return err
		}

		f.SetFloat(fl)
		return nil
	}
}

func sliceSetter(f reflect.Value, typ reflect.Type) setter {
	return func(s string) error {
		sl := reflect.MakeSlice(typ, 0, 0)

		if typ.Elem().Kind() == reflect.Uint8 {
			sl = reflect.ValueOf([]byte(s))
		} else if len(strings.TrimSpace(s)) > 0 {
			vals := strings.Split(s, ",")
			sl = reflect.MakeSlice(typ, len(vals), len(vals))
			for i, val := range vals {
				fs, err := getSetter(sl.Index(i))
				if err != nil {
					return err
				}

				if err := fs(val); err != nil {
					return err
				}
			}
		}

		f.Set(sl)
		return nil
	}
}
