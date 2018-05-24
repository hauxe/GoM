package crudl

import (
	"context"
	"reflect"

	"github.com/pkg/errors"

	gomHTTP "github.com/hauxe/GoM/http"
)

func validatePrimaryKey(pk string) gomHTTP.ParamValidator {
	return func(_ context.Context, obj interface{}) error {
		m, ok := obj.(map[string]interface{})
		if !ok {
			return errors.Errorf("invalid type %s of parameter", reflect.TypeOf(obj).String())
		}
		if _, ok = m[pk]; !ok {
			return errors.Errorf("missing primary key %s", pk)
		}
		return nil
	}
}

func getMethodValidator(method string, validator Validator) gomHTTP.ParamValidator {
	return func(_ context.Context, obj interface{}) error {
		return validator(method, obj)
	}
}
