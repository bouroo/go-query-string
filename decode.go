package query

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Unmarshal(data string, v any) error {
	d := newDecoder(data, v)
	return d.Do()
}

type decode struct {
	data string
	obj  any
}

func newDecoder(data string, v any) *decode {
	return &decode{
		data: data,
		obj:  v,
	}
}

type decodeError struct{ error }

func (d *decode) error(err error) {
	panic(decodeError{err})
}

func (d *decode) Do() (err error) {
	defer func() {
		if r := recover(); r != nil {
			if je, ok := r.(decodeError); ok {
				err = je.error
			} else {
				panic(r)
			}
		}
	}()

	if reflect.ValueOf(d.obj).Kind() != reflect.Pointer {
		d.error(fmt.Errorf("%v is non-pointer", reflect.TypeOf(d.obj)))
	}

	data := d.splitSeperate()
	d.mapValueToObj(data, reflect.ValueOf(d.obj))
	return
}

func (d *decode) splitSeperate() map[string]string {
	m := make(map[string]string)
	queryParts := strings.Split(d.data, seperator)
	for _, query := range queryParts {
		sepData := strings.SplitN(query, equal, 2)
		if len(sepData) < 2 {
			d.error(fmt.Errorf("%v is an invalid query", sepData))
		} else {
			m[sepData[0]] = sepData[1]
		}
	}
	return m
}

func (d *decode) getName(sf reflect.StructField) string {
	name := sf.Tag.Get(tag)
	if name == tagNameFollowType {
		name = sf.Name
	} else if name == "" {
		name = convertToSnakeCase(sf.Name)
	}
	return name
}

func (e *decode) ignoreField(v reflect.StructField) bool {
	return v.Tag.Get(tag) == tagIgnore
}

func (d *decode) mapValueToObj(data map[string]string, v reflect.Value) {
	elem := v.Elem()
	elemType := elem.Type()

	for i := 0; i < elem.NumField(); i++ {
		fieldValue := elem.Field(i)
		fieldType := elemType.Field(i)

		if !fieldValue.CanSet() {
			d.error(fmt.Errorf("%v is unexported field", fieldType.Name))
		}

		if d.ignoreField(fieldType) {
			continue
		}

		fieldName := d.getName(fieldType)
		if fieldValue.CanAddr() {
			if dataValue, ok := data[fieldName]; ok {
				d.valueToString(fieldName, dataValue, fieldValue)
			}
		}
	}
}

func (d *decode) valueToString(fieldName, data string, v reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsZero() || v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		d.valueToString(fieldName, data, v.Elem())
	case reflect.String:
		v.SetString(data)
	case reflect.Bool:
		value, err := strconv.ParseBool(data)
		if err != nil {
			d.error(fmt.Errorf("err: %v, %v=%v", err, fieldName, data))
		}
		v.SetBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(data, 10, v.Type().Bits())
		if err != nil {
			d.error(fmt.Errorf("err: %v, %v=%v", err, fieldName, data))
		}
		v.SetInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, err := strconv.ParseUint(data, 10, v.Type().Bits())
		if err != nil {
			d.error(fmt.Errorf("err: %v, %v=%v", err, fieldName, data))
		}
		v.SetUint(value)
	case reflect.Float32, reflect.Float64:
		value, err := strconv.ParseFloat(data, v.Type().Bits())
		if err != nil {
			d.error(fmt.Errorf("err: %v, %v=%v", err, fieldName, data))
		}
		v.SetFloat(value)
	default:
		d.error(fmt.Errorf("type does not support"))
	}
}
