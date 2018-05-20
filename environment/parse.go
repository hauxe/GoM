package environment

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	lib "github.com/hauxe/gom/library"
	"github.com/pkg/errors"
)

// CreateENVOptions create env options
type CreateENVOptions func(*ENVConfig) error

const (
	envKey = "ENV"
	envTag = "env"
)

// Environment types
const (
	Development = "development"
	Testing     = "testing"
	Staging     = "staging"
	Production  = "production"
)

// Environment returns the current running environment
func Environment() string {
	environment := os.Getenv(envKey)
	if environment == Production || environment == Staging || environment == Testing {
		return environment
	}
	return Development
}

// ENVConfig defines environment configs
type ENVConfig struct {
	Prefix string
}

// ENV defines environment object
type ENV struct {
	Config *ENVConfig
}

// CreateENV create environment object
func CreateENV(options ...CreateENVOptions) (*ENV, error) {
	config := ENVConfig{}
	for _, op := range options {
		if err := op(&config); err != nil {
			return nil, errors.Wrap(err, lib.StringTags("create env", "option error"))
		}
	}
	return &ENV{Config: &config}, nil
}

// SetPrefixOption set environment prefix option
func SetPrefixOption(prefix string) CreateENVOptions {
	return func(config *ENVConfig) error {
		config.Prefix = prefix
		return nil
	}
}

// EVString gets environment variable by key, returns its value as string
// or returns fallback if not available
func (e *ENV) EVString(key string, fallback string) string {
	value, found := os.LookupEnv(key)
	if !found {
		return fallback
	}
	return value
}

// EVInt64 gets environment variable by key, returns its value as int64
// or returns fallback if not available
func (e *ENV) EVInt64(key string, fallback int64) (int64, error) {
	value, found := os.LookupEnv(key)
	if !found {
		return fallback, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

// EVInt gets environment variable by key, returns its value as int
// or returns fallback if not available
func (e *ENV) EVInt(key string, fallback int) (int, error) {
	value, found := os.LookupEnv(key)
	if !found {
		return fallback, nil
	}
	return strconv.Atoi(value)
}

// EVUInt64 gets environment variable by key, returns its value as uint64
// or returns fallback if not available
func (e *ENV) EVUInt64(key string, fallback uint64) (uint64, error) {
	value, found := os.LookupEnv(key)
	if !found {
		return fallback, nil
	}
	return strconv.ParseUint(value, 10, 64)
}

// EVBool gets environment variable by key, returns its value as bool
// or returns fallback if not available
func (e *ENV) EVBool(key string, fallback bool) (bool, error) {
	value, found := os.LookupEnv(key)
	if !found {
		return fallback, nil
	}
	return strconv.ParseBool(value)
}

// Parse parses environment variables to struct
func (e *ENV) Parse(obj interface{}, validators ...func(interface{}) error) (err error) {
	defer lib.Recover(func(er error) {
		if er != nil {
			err = er
		}
	})
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr {
		return errors.New("object type is not pointer")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return errors.Errorf("object type is not struct <%s>", rv.Kind().String())
	}
	if err = e.scanStructENV(rv); err != nil {
		return err
	}
	for _, validator := range validators {
		if validator != nil {
			if err = validator(obj); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *ENV) scanStructENV(rv reflect.Value) (err error) {
	var wg sync.WaitGroup
	var pointerErr error
	var structErr error
	prefix := ""
	if e.Config != nil && e.Config.Prefix != "" {
		prefix = "_" + e.Config.Prefix
	}
	for i := 0; i < rv.NumField(); i++ {
		if rv.Field(i).Kind() == reflect.Ptr && rv.Field(i).Elem().Kind() == reflect.Struct {
			wg.Add(1)
			go func(rv reflect.Value) {
				pointerErr = e.scanStructENV(rv)
				wg.Done()
			}(rv.Field(i).Elem())
		}
		if rv.Field(i).Kind() == reflect.Struct {
			wg.Add(1)
			go func(rv reflect.Value) {
				structErr = e.scanStructENV(rv)
				wg.Done()
			}(rv.Field(i))
		}
		tag, ok := rv.Type().Field(i).Tag.Lookup(envTag)
		if !ok {
			continue
		}
		tags := strings.Split(tag, ",")
		n := len(tags)
		// was this field marked for skipping?
		if n == 0 || tags[0] == "-" {
			continue
		}
		if err = e.getFieldENV(rv.Type().Field(i).Name, rv.Field(i), prefix+tags[0]); err != nil {
			return errors.Wrap(err, lib.StringTags("scan struct env"))
		}
	}
	wg.Wait()
	if pointerErr != nil {
		return pointerErr
	}
	if structErr != nil {
		return structErr
	}
	return
}

func (e *ENV) getFieldENV(fieldName string, field reflect.Value, name string) error {
	if name == "" {
		return nil
	}
	s, found := os.LookupEnv(name)
	if !found {
		return nil
	}
	if !field.CanSet() {
		return errors.Errorf("field %s with type %s cant be set", fieldName,
			field.Kind())
	}
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64, err := strconv.ParseInt(s, 10, field.Type().Bits())
		if err != nil {
			return errors.Errorf("field %s convert environment %s to %s failed",
				fieldName, s, field.Kind().String())
		}
		field.SetInt(i64)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u64, err := strconv.ParseUint(s, 10, field.Type().Bits())
		if err != nil {
			return errors.Errorf("field %s convert environment %s to %s failed",
				fieldName, s, field.Kind().String())
		}
		field.SetUint(u64)
		return nil
	case reflect.Float32, reflect.Float64:
		f64, err := strconv.ParseFloat(s, field.Type().Bits())
		if err != nil {
			return errors.Errorf("field %s convert environment %s to %s failed",
				fieldName, s, field.Kind().String())
		}
		field.SetFloat(f64)
		return nil
	case reflect.String:
		field.SetString(s)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return errors.Errorf("field %s convert environment %s to %s failed",
				fieldName, s, field.Kind().String())
		}
		field.SetBool(b)
		return nil
	}
	return errors.Errorf("convert field %s, type %s not supported",
		fieldName, field.Kind().String())
}
