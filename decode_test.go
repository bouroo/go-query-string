package query_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/teerapon19/go-query-string"
)

func TestUnmarshal(t *testing.T) {

	equal := func(t *testing.T, expect interface{}, actual interface{}) {
		if !reflect.DeepEqual(expect, actual) {
			t.Fatalf("expect: %v, actual: %v", expect, actual)
		}
	}

	t.Run("type pointer int", func(t *testing.T) {

		type QueryParams struct {
			Number *int `query:"number"`
		}

		var actual QueryParams

		number := 1
		expect := QueryParams{
			Number: &number,
		}
		query.Unmarshal("number=1", &actual)

		equal(t, expect, actual)
	})

	t.Run("type pointer string", func(t *testing.T) {

		type QueryParams struct {
			Text *string `query:"text"`
		}

		var actual QueryParams

		number := "hello"
		expect := QueryParams{
			Text: &number,
		}
		query.Unmarshal("text=hello", &actual)

		equal(t, expect, actual)
	})

	t.Run("type pointer string not found need nil", func(t *testing.T) {

		type QueryParams struct {
			Text *string `query:"text"`
		}

		var actual QueryParams

		expect := QueryParams{
			Text: nil,
		}
		query.Unmarshal("text2=hello", &actual)

		equal(t, expect, actual)
	})

	t.Run("type string", func(t *testing.T) {

		type QueryParams struct {
			Text string `query:"text"`
		}

		var actual QueryParams

		expect := QueryParams{
			Text: "hello",
		}
		query.Unmarshal("text=hello", &actual)

		equal(t, expect, actual)
	})

	t.Run("type bool", func(t *testing.T) {

		type QueryParams struct {
			IsTrue bool `query:"is_true"`
		}

		var actual QueryParams

		expect := QueryParams{
			IsTrue: true,
		}
		query.Unmarshal("is_true=true", &actual)

		equal(t, expect, actual)
	})

	t.Run("type int int8 int16 int32 int64 and multiple value", func(t *testing.T) {

		type QueryParams struct {
			Int   int   `query:"int"`
			Int8  int8  `query:"int8"`
			Int16 int16 `query:"int16"`
			Int32 int32 `query:"int32"`
			Int64 int64 `query:"int64"`
		}

		var actual QueryParams

		expect := QueryParams{
			Int:   1,
			Int8:  2,
			Int16: 3,
			Int32: 4,
			Int64: 5,
		}
		err := query.Unmarshal("int=1&int8=2&int16=3&int32=4&int64=5", &actual)
		if err != nil {
			t.Fatal(err)
		}

		equal(t, expect, actual)
	})

	t.Run("type float32 float64 and multiple value", func(t *testing.T) {

		type QueryParams struct {
			Float32 float32 `query:"float32"`
			Float64 float64 `query:"float64"`
		}

		var actual QueryParams

		expect := QueryParams{
			Float32: 9.123456,
			Float64: 9.1234567890123,
		}
		err := query.Unmarshal("float32=9.123456&float64=9.1234567890123", &actual)
		if err != nil {
			t.Fatal(err)
		}

		equal(t, expect, actual)
	})

	t.Run("without tag", func(t *testing.T) {

		type QueryParams struct {
			IsTrue bool
		}

		var actual QueryParams

		expect := QueryParams{
			IsTrue: true,
		}
		err := query.Unmarshal("is_true=true", &actual)
		if err != nil {
			t.Fatal(err)
		}

		equal(t, expect, actual)
	})

	t.Run("with tag case:normal", func(t *testing.T) {

		type QueryParams struct {
			IsTrue bool `query:"name:type"`
		}

		var actual QueryParams

		expect := QueryParams{
			IsTrue: true,
		}
		err := query.Unmarshal("IsTrue=true", &actual)
		if err != nil {
			t.Fatal(err)
		}

		equal(t, expect, actual)
	})

	t.Run("with mix", func(t *testing.T) {

		type QueryParams struct {
			Text       string `query:"text"`
			ID         string
			Number     int64
			Vat        float64
			MiddleName *string
		}

		var actual QueryParams

		expect := QueryParams{
			Text:       "TEXT",
			ID:         "1234567",
			Number:     0,
			Vat:        1.1,
			MiddleName: nil,
		}

		err := query.Unmarshal("text=TEXT&id=1234567&number=0&vat=1.1", &actual)
		if err != nil {
			t.Fatal(err)
		}

		equal(t, expect, actual)
	})

	t.Run("with unexported field return error", func(t *testing.T) {

		type QueryParams struct {
			text string `query:"text"`
		}

		var actual QueryParams

		expectErr := fmt.Errorf("%v is unexported field", reflect.ValueOf(actual).Type().Field(0).Name)

		actualErr := query.Unmarshal("text=TEXT", &actual)

		equal(t, actual.text, "")
		equal(t, expectErr, actualErr)
	})

	t.Run("without ignore", func(t *testing.T) {

		type QueryParams struct {
			Ignore string `query:"-"`
			IsTrue bool
		}

		var actual QueryParams

		expect := QueryParams{
			IsTrue: true,
		}
		err := query.Unmarshal("is_true=true", &actual)
		if err != nil {
			t.Fatal(err)
		}

		equal(t, expect, actual)
	})

	t.Run("with mix and ignore", func(t *testing.T) {

		type QueryParams struct {
			Text       string `query:"text"`
			ID         string `query:"-"`
			Number     int64
			Vat        float64 `query:"-"`
			MiddleName *string
		}

		var actual QueryParams

		expect := QueryParams{
			Text:       "TEXT",
			ID:         "",
			Number:     888,
			Vat:        0,
			MiddleName: nil,
		}

		err := query.Unmarshal("text=TEXT&number=888", &actual)
		if err != nil {
			t.Fatal(err)
		}

		equal(t, expect, actual)
	})

}

func BenchmarkUnmarshal(b *testing.B) {
	type QueryParams struct {
		Text       string `query:"text"`
		ID         string
		Number     int64
		Vat        float64
		MiddleName *string
	}

	b.ReportAllocs()
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var actual QueryParams
			query.Unmarshal("text=TEXT&id=1234567&number=0&vat=1.1", &actual)
		}
	})
}

func BenchmarkDecoderPool(b *testing.B) {
	type QueryParams struct {
		Text       string `query:"text"`
		ID         string
		Number     int64
		Vat        float64
		MiddleName *string
	}

	dp := &query.DecoderPool{}

	b.ReportAllocs()
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var actual QueryParams
			decoder := dp.Get("text=TEXT&id=1234567&number=0&vat=1.1", &actual)
			defer dp.Put(decoder)
			_ = decoder.Do()
		}
	})
}
