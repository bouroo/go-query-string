package query

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

func Marshal(v any) (string, error) {
	e := newEncoder(v)
	return e.do()
}

type encode struct {
	qb  bytes.Buffer
	obj reflect.Value
}

func newEncoder(obj any) *encode {
	vObj := reflect.ValueOf(obj)

	if reflect.TypeOf(obj).Kind() == reflect.Pointer {
		vObj = vObj.Elem()
	}

	return &encode{
		qb:  bytes.Buffer{},
		obj: vObj,
	}
}

type encodeError struct{ error }

func (e *encode) error(err error) {
	panic(encodeError{err})
}

func (e *encode) do() (query string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if je, ok := r.(encodeError); ok {
				err = je.error
			} else {
				panic(r)
			}
		}
	}()

	for i := 0; i < e.obj.NumField(); i++ {
		e.pair(e.obj.Type().Field(i), e.obj.Field(i))
		if i != e.obj.NumField()-1 {
			e.qb.WriteString(seperator)
		}
	}

	query = e.qb.String()
	return
}

func (e *encode) pair(fs reflect.StructField, v reflect.Value) {
	bb := bytes.Buffer{}
	key := e.key(fs)
	bb.WriteString(key)
	bb.WriteString(equal)
	bb.WriteString(e.valueToString(v))
	e.qb.Write(bb.Bytes())
}

func (e *encode) key(v reflect.StructField) string {
	name := v.Tag.Get(tag)

	if name == tagNameFollowType {
		return v.Name
	}

	return ConvertToSnakeCase(v.Name)
}

func (e *encode) valueToString(v reflect.Value) (s string) {

	// reflect pointer
	if v.Kind() == reflect.Pointer {
		s = e.valueToString(v.Elem())
		return
	}

	switch v.Kind() {
	case reflect.String:
		s = v.String()
	case reflect.Bool:
		s = strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s = strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s = strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		s = strconv.FormatFloat(v.Float(), 'f', -1, v.Type().Bits())
	default:
		e.error(fmt.Errorf("type does not support"))
	}

	return
}
