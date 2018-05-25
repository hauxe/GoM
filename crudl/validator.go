package crudl

import (
	"context"
	"reflect"

	"github.com/pkg/errors"

	gomHTTP "github.com/hauxe/gom/http"
)

func validatePrimaryKey(pk int) gomHTTP.ParamValidator {
	return func(_ context.Context, obj interface{}) error {
		rv := reflect.ValueOf(obj)
		rv = reflect.Indirect(rv)
		pk := rv.Field(pk)
		if !pk.CanInterface() {
			return errors.Errorf("missing or invalid primary key")
		}
		switch pk.Kind() {
		case reflect.Int64:
			if pk.Int() == 0 {
				return errors.Errorf("missing primary key")
			}
		case reflect.String:
			if pk.String() == "" {
				return errors.Errorf("missing primary key")
			}
		}
		return nil
	}
}

func getMethodValidator(method string, validator Validator) gomHTTP.ParamValidator {
	return func(_ context.Context, obj interface{}) error {
		return validator(method, obj)
	}
}
