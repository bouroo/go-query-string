package query

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Unmarshal(data string, v any) error {
	d := newDecoder(data, v)
	return d.do()
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

func (d *decode) do() (err error) {
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

func (d *decode) splitSeperate() (m map[string]string) {
	m = map[string]string{}
	for _, v := range strings.Split(d.data, seperator) {
		sepData := strings.SplitN(v, equal, 2)
		if len(sepData) < 2 {
			d.error(fmt.Errorf("%v is invalid query", sepData))
		}
		m[sepData[0]] = sepData[1]
	}
	return
}

func (d *decode) getName(sf reflect.StructField) string {
	name := sf.Tag.Get(tag)
	if name == tagNameFollowType {
		name = sf.Name
	} else if name == "" {
		name = ConvertToSnakeCase(sf.Name)
	}
	return name
}

func (d *decode) mapValueToObj(data map[string]string, v reflect.Value) {
	elem := v.Elem()
	for i := 0; i < elem.NumField(); i++ {
		fv := elem.Field(i)
		ft := elem.Type().Field(i)

		if !fv.CanSet() {
			d.error(fmt.Errorf("%v is unexported field", ft.Name))
		}

		fieldName := d.getName(ft)
		if data, ok := data[fieldName]; ok {
			d.valueToString(fieldName, data, fv)
		}
	}

}

func (d *decode) valueToString(fieldName, data string, v reflect.Value) {

	// reflect pointer
	if v.Kind() == reflect.Pointer {
		if v.IsZero() || v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		d.valueToString(fieldName, data, v.Elem())
		return
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(data)
	case reflect.Bool:
		value, err := strconv.ParseBool(data)
		if err != nil {
			d.error(fmt.Errorf("err: %v, %v=%v", err, fieldName, data))
		}
		v.SetBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(data, 10, 64)
		if err != nil {
			d.error(fmt.Errorf("err: %v, %v=%v", err, fieldName, data))
		}
		v.SetInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, err := strconv.ParseUint(data, 10, 64)
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