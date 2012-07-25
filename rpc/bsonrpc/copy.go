package bsonrpc

import (
	"errors"
	"fmt"
	"launchpad.net/mgo/v2/bson"
	"reflect"
)

const FTrace = false

func CopyTo(src bson.M, dst interface{}) (err error) {
	if FTrace {
		fmt.Println("+CopyTo()")
		defer fmt.Println("-CopyTo()")
	}
	dval := reflect.ValueOf(dst)
	dtyp := dval.Type()

	if dtyp.Kind() == reflect.Ptr {
		dval = dval.Elem()
	}

	err = copyMapToVal(src, dval)

	return
}

func copyMapToVal(src bson.M, dval reflect.Value) (err error) {
	if FTrace {
		fmt.Println("+copyMapToVal()")
		defer fmt.Println("-copyMapToVal()")
	}
	switch dval.Type().Kind() {
	case reflect.Map:
		return copyMapToMapVal(src, dval)
	case reflect.Struct:
		return copyMapToStructVal(src, dval)
	case reflect.Interface:
		// if we're copying into an interface, just set it and forget it
		dval.Set(reflect.ValueOf(&src).Elem())
		return
	default:
		fmt.Printf("Kind is %v\n", dval.Type().Kind())
	}

	return
}

func copyMapToMapVal(src bson.M, dval reflect.Value) (err error) {
	if FTrace {
		fmt.Println("+copyMapToMapVal()")
		defer fmt.Println("-copyMapToMapVal()")
	}
	if dval.IsNil() {
		dval.Set(reflect.MakeMap(dval.Type()))
	}
	elemType := dval.Type().Elem()
	for key, val := range src {
		eval := reflect.New(elemType).Elem()
		err = copyValToVal(reflect.ValueOf(val), eval)
		if err != nil {
			return
		}
		dval.SetMapIndex(reflect.ValueOf(key), eval)
	}

	return
}

func copyMapToStructVal(src bson.M, dval reflect.Value) (err error) {
	if FTrace {
		fmt.Println("+copyMapToStructVal()")
		defer fmt.Println("-copyMapToStructVal()")
	}
	for key, val := range src {
		eval := dval.FieldByName(key)
		err = copyValToVal(reflect.ValueOf(val), eval)
		if err != nil {
			return
		}
	}

	return
}

func copyValToVal(sval, dval reflect.Value) (err error) {
	if FTrace {
		fmt.Println("+copyValToVal()")
		defer fmt.Println("-copyValToVal()")
	}
	switch sval.Type().Kind() {
	case reflect.Slice:
		if dval.Type().Kind() != reflect.Slice {
			err = errors.New("Source and destination fields don't match type")
			return
		}
		err = copySliceValToSliceVal(sval, dval)
		return
	case reflect.Interface:
		err = copyValToVal(sval.Elem(), dval)
		return
	case reflect.Map:
		src := sval.Interface().(bson.M)
		copyMapToVal(src, dval)
	default:
		defer func() {
			e := recover()
			if e != nil {
				err = errors.New(fmt.Sprintf("Could not assign: %v", e))
				fmt.Println(err)
			}
		}()
		dval.Set(sval)
	}
	return
}

func copySliceValToSliceVal(sval, dval reflect.Value) (err error) {
	if FTrace {
		fmt.Println("+copySliceValToSliceVal()")
		defer fmt.Println("-copySliceValToSliceVal()")
	}
	length := sval.Len()
	dval.Set(reflect.MakeSlice(dval.Type(), length, length))
	for i := 0; i < length; i++ {
		copyValToVal(sval.Index(i), dval.Index(i))
	}
	return
}
