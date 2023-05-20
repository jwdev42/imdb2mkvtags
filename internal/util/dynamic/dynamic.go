//This file is part of imdb2mkvtags ©2021 Jörg Walter

package dynamic

import (
	"fmt"
	"reflect"
)

// assigns data to field name of struct rec
func SetStructField(name string, rec, data interface{}) {
	val := reflect.Indirect(reflect.ValueOf(rec))
	fieldVal := val.FieldByName(name)
	if fieldVal == (reflect.Value{}) {
		panic(fmt.Errorf(`Field "%s" not a member of struct %s`, name, val.Type().Name()))
	}
	fieldVal.Set(reflect.ValueOf(data))
}

func SetStructFieldCallback(name string, rec, callback interface{}) error {
	vf := reflect.ValueOf(callback)

	//verification of callback
	if vf.Kind() != reflect.Func {
		panic("callback must be a function")
	}
	if vf.Type().NumOut() != 2 {
		panic("callback must return 2 values")
	}
	if n := vf.Type().Out(1).Name(); n != "error" {
		panic(fmt.Errorf("callback's 2nd return value must be of type \"error\", but is \"%s\"", n))
	}

	rv := vf.Call(nil)

	if !rv[1].IsNil() { //Function call returned an Error
		err := rv[1].Interface().(error)
		return err
	}
	payload := rv[0].Interface()
	SetStructField(name, rec, payload)
	return nil
}

// Returns all fields of rec that have a non-empty struct tag of the given key and all corresponding tag values.
// Returns nil, nil if no such fields were found. "rec" must be a pointer to a struct.
func FieldsByStructTag(key string, rec interface{}) ([]reflect.Value, []string) {
	fields := make([]reflect.Value, 0, 10)
	tags := make([]string, 0, 10)

	structVal := reflect.Indirect(reflect.ValueOf(rec))
	for i := 0; i < structVal.NumField(); i++ {
		tag := structVal.Type().Field(i).Tag.Get(key)
		if tag == "" {
			continue
		}
		fields = append(fields, structVal.Field(i))
		tags = append(tags, tag)
	}
	if len(fields) < 1 {
		return nil, nil
	}
	return fields, tags
}

func mustBeType(i interface{}, t reflect.Kind) {
	if i == nil || reflect.TypeOf(i).Kind() != t {
		panic(fmt.Errorf("Assertion failed: Type is not %s", t.String()))
	}
}
