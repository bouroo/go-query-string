package query

import (
	"bytes"
	"reflect"
	"sync"
)

type DecoderPool struct {
	pool sync.Pool
}

func (dp *DecoderPool) Get(data string, v any) *decode {
	p := dp.pool.Get()
	if p == nil {
		return &decode{
			data: data,
			obj:  v,
		}
	}
	return p.(*decode)
}

func (dp *DecoderPool) Put(d *decode) {
	dp.pool.Put(d)
}

type EncoderPool struct {
	pool sync.Pool
}

func (ep *EncoderPool) Get(v any) *encode {
	p := ep.pool.Get()
	if p == nil {
		vObj := reflect.ValueOf(v)

		if reflect.TypeOf(v).Kind() == reflect.Pointer {
			vObj = vObj.Elem()
		}
		return &encode{
			qb:  bytes.Buffer{},
			obj: vObj,
		}
	}
	return p.(*encode)
}

func (ep *EncoderPool) Put(e *encode) {
	ep.pool.Put(e)
}
