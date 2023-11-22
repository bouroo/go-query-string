package query

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

func Marshal(v any) (string, error) {
	e := newEncoder(v)
	return e.Do()
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

func (e *encode) Do() (query string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if je, ok := r.(encodeError); ok {
				err = je.error
			} else {
				panic(r)
			}
		}
	}()

	isFrist := true
	for i := 0; i < e.obj.NumField(); i++ {
		fieldType := e.obj.Type().Field(i)
		if e.ignoreField(fieldType) {
			continue
		}
		if !isFrist {
			e.qb.WriteString(seperator)
		}
		e.pair(fieldType, e.obj.Field(i))
		isFrist = false
	}

	query = e.qb.String()
	return
}

func (e *encode) ignoreField(v reflect.StructField) bool {
	return v.Tag.Get(tag) == tagIgnore
}

func (e *encode) pair(fs reflect.StructField, v reflect.Value) {
	key := e.key(fs)
	value := e.valueToString(v)
	e.qb.WriteString(key)
	e.qb.WriteString(equal)
	e.qb.WriteString(value)
}

func (e *encode) key(v reflect.StructField) string {
	name := v.Tag.Get(tag)

	if name == tagNameFollowType {
		return v.Name
	}

	return convertToSnakeCase(v.Name)
}

func (e *encode) valueToString(v reflect.Value) (s string) {

	switch v.Kind() {
	case reflect.Pointer:
		s = e.valueToString(v.Elem())
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
