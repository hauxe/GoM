package environment

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/pkg/errors"

	lib "github.com/hauxe/gom/library"
	"github.com/stretchr/testify/require"
)

var env, _ = CreateENV()

func TestEnvironment(t *testing.T) {
	// no parallel
	require.Equal(t, Development, Environment())
	environmentList := []string{Production, Staging, Testing, Development}
	for _, environment := range environmentList {
		os.Setenv(envKey, environment)
		require.Equal(t, environment, Environment())
	}
}

func TestEVString(t *testing.T) {
	t.Parallel()
	key := "LIB_NAME"
	val := "tres lib"
	os.Setenv(key, val)
	testCases := []struct {
		Name     string
		Key      string
		Fallback string
		Value    string
	}{
		{"not_exist", "NOT_EXIST_KEY", "none", "none"},
		{"exist", key, "none", val},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			value := env.EVString(tc.Key, tc.Fallback)
			require.Equal(t, tc.Value, value)
		})
	}
}

func TestEVInt64(t *testing.T) {
	t.Parallel()
	key := "LIB_INT64"
	val := int64(999999999999999)
	os.Setenv(key, fmt.Sprintf("%d", val))
	testCases := []struct {
		Name     string
		Key      string
		Fallback int64
		Value    int64
	}{
		{"not_exist", "NOT_EXIST_KEY", -1, -1},
		{"exist", key, -1, val},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			value, err := env.EVInt64(tc.Key, tc.Fallback)
			require.Nil(t, err)
			require.Equal(t, tc.Value, value)
		})
	}
}

func TestEVInt(t *testing.T) {
	t.Parallel()
	key := "LIB_INT"
	val := -999
	os.Setenv(key, fmt.Sprintf("%d", val))
	testCases := []struct {
		Name     string
		Key      string
		Fallback int
		Value    int
	}{
		{"not_exist", "NOT_EXIST_KEY", -1, -1},
		{"exist", key, -1, val},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			value, err := env.EVInt(tc.Key, tc.Fallback)
			require.Nil(t, err)
			require.Equal(t, tc.Value, value)
		})
	}
}

func TestEVUInt64(t *testing.T) {
	t.Parallel()
	key := "LIB_UINT64"
	val := uint64(999)
	os.Setenv(key, fmt.Sprintf("%d", val))
	testCases := []struct {
		Name     string
		Key      string
		Fallback uint64
		Value    uint64
	}{
		{"not_exist", "NOT_EXIST_KEY", 0, 0},
		{"exist", key, 0, val},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			value, err := env.EVUInt64(tc.Key, tc.Fallback)
			require.Nil(t, err)
			require.Equal(t, tc.Value, value)
		})
	}
}

func TestEVBool(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name     string
		Key      string
		Val      string
		Fallback bool
		Value    bool
		Error    bool
	}{
		{"not_exist", "NOT_EXIST_KEY", "true", false, false, false},
		{"invalid", "INVALID", "invalid", false, false, true},
		{"true", "LIB_BOOL_1", "true", false, true, false},
		{"false", "LIB_BOOL_2", "false", true, false, false},
		{"True", "LIB_BOOL_3", "True", false, true, false},
		{"False", "LIB_BOOL_4", "False", false, false, false},
		{"TRUE", "LIB_BOOL_5", "TRUE", true, true, false},
		{"FALSE", "LIB_BOOL_6", "FALSE", true, false, false},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			if tc.Name != "not_exist" {
				os.Setenv(tc.Key, tc.Val)
			}
			value, err := env.EVBool(tc.Key, tc.Fallback)
			if tc.Error {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.Value, value)
			}
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()
	type validSubStruct struct {
		A1 int     `env:"PA1,validator=a1"`
		B1 int64   `env:"PB1,validator=b1"`
		C1 uint    `env:"PC1,validator=c1"`
		D1 uint64  `env:"PD1,validator=d1"`
		E1 float64 `env:"PE1,validator=e1"`
		F1 string  `env:"PF1,validator=f1"`
		G1 bool    `env:"PG1,validator=g1"`
	}
	type validSubStruct1 struct {
		A2 int     `env:"PA2,validator=a2"`
		B2 int64   `env:"PB2,validator=b2"`
		C2 uint    `env:"PC2,validator=c2"`
		D2 uint64  `env:"PD2,validator=d2"`
		E2 float64 `env:"PE2,validator=e2"`
		F2 string  `env:"PF2,validator=f2"`
		G2 bool    `env:"PG2,validator=g2"`
	}
	type validStruct struct {
		A int     `env:"PA,validator=a"`
		B int64   `env:"PB,validator=b"`
		C uint    `env:"PC,validator=c"`
		D uint64  `env:"PD,validator=d"`
		E float64 `env:"PE,validator=e"`
		F string  `env:"PF,validator=f"`
		G bool    `env:"PG,validator=g"`
		H *validSubStruct
		I validSubStruct1
	}
	t.Run("error_not_pointer", func(t *testing.T) {
		t.Parallel()
		data := validStruct{
			A: 1,
		}
		err := env.Parse(data)
		require.Error(t, err)
		require.Equal(t, 1, data.A)
	})
	t.Run("error_not_truct", func(t *testing.T) {
		t.Parallel()
		data := 1
		err := env.Parse(&data)
		require.Error(t, err)
		require.Equal(t, 1, data)
	})
	t.Run("error_scan_struct", func(t *testing.T) {
		t.Parallel()
		data := struct {
			A int `env:"TA"`
		}{
			A: 1,
		}
		os.Setenv("TA", "invalid")
		err := env.Parse(&data, func(interface{}) error {
			return errors.New("test failed scan struct")
		})
		require.Error(t, err)
		require.Equal(t, 1, data.A)
	})
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		data := validStruct{
			A: -1,
			B: -10,
			C: 1,
			D: 1000,
			E: 1.8,
			F: "test",
			G: true,
			H: &validSubStruct{
				A1: -2,
				B1: -20,
				C1: 2,
				D1: 2000,
				E1: 2.8,
				F1: "2test",
				G1: true,
			},
			I: validSubStruct1{
				A2: -3,
				B2: -30,
				C2: 3,
				D2: 3000,
				E2: 3.8,
				F2: "3test",
				G2: true,
			},
		}
		clone := validStruct{
			A: -1,
			B: -10,
			C: 1,
			D: 1000,
			E: 1.8,
			F: "test",
			G: true,
			H: &validSubStruct{
				A1: -2,
				B1: -20,
				C1: 2,
				D1: 2000,
				E1: 2.8,
				F1: "2test",
				G1: true,
			},
			I: validSubStruct1{
				A2: -3,
				B2: -30,
				C2: 3,
				D2: 3000,
				E2: 3.8,
				F2: "3test",
				G2: true,
			},
		}
		os.Setenv("PA", lib.ToString(data.A+1111))
		os.Setenv("PB", lib.ToString(data.B+1111))
		os.Setenv("PC", lib.ToString(data.C+1111))
		os.Setenv("PD", lib.ToString(data.D+1111))
		os.Setenv("PE", lib.ToString(data.E+1111))
		os.Setenv("PF", data.F+lib.ToString(1111))
		os.Setenv("PG", lib.ToString(!data.G))
		os.Setenv("PA1", lib.ToString(data.H.A1+1111))
		os.Setenv("PB1", lib.ToString(data.H.B1+1111))
		os.Setenv("PC1", lib.ToString(data.H.C1+1111))
		os.Setenv("PD1", lib.ToString(data.H.D1+1111))
		os.Setenv("PE1", lib.ToString(data.H.E1+1111))
		os.Setenv("PF1", data.H.F1+lib.ToString(1111))
		os.Setenv("PG1", lib.ToString(!data.H.G1))
		os.Setenv("PA2", lib.ToString(data.I.A2+1111))
		os.Setenv("PB2", lib.ToString(data.I.B2+1111))
		os.Setenv("PC2", lib.ToString(data.I.C2+1111))
		os.Setenv("PD2", lib.ToString(data.I.D2+1111))
		os.Setenv("PE2", lib.ToString(data.I.E2+1111))
		os.Setenv("PF2", data.I.F2+lib.ToString(1111))
		os.Setenv("PG2", lib.ToString(!data.I.G2))
		err := env.Parse(&data,
			func(v interface{}) error {
				obj, ok := v.(*validStruct)
				if !ok {
					return errors.New("can convert back to object")
				}
				obj.A += 999
				return nil
			},
			func(v interface{}) error {
				obj, ok := v.(*validStruct)
				if !ok {
					return errors.New("can convert back to object")
				}
				obj.B += 999
				return nil
			},
			func(v interface{}) error {
				obj, ok := v.(*validStruct)
				if !ok {
					return errors.New("can convert back to object")
				}
				obj.C += 999
				return nil
			})
		require.Nil(t, err)
		require.Equal(t, clone.A+1111+999, data.A)
		require.Equal(t, clone.B+1111+999, data.B)
		require.Equal(t, clone.C+1111+999, data.C)
		require.Equal(t, clone.D+1111, data.D)
		require.Equal(t, clone.E+1111, data.E)
		require.Equal(t, clone.F+"1111", data.F)
		require.Equal(t, !clone.G, data.G)
		require.Equal(t, clone.H.A1+1111, data.H.A1)
		require.Equal(t, clone.H.B1+1111, data.H.B1)
		require.Equal(t, clone.H.C1+1111, data.H.C1)
		require.Equal(t, clone.H.D1+1111, data.H.D1)
		require.Equal(t, clone.H.E1+1111, data.H.E1)
		require.Equal(t, clone.H.F1+"1111", data.H.F1)
		require.Equal(t, !clone.H.G1, data.H.G1)
		require.Equal(t, clone.I.A2+1111, data.I.A2)
		require.Equal(t, clone.I.B2+1111, data.I.B2)
		require.Equal(t, clone.I.C2+1111, data.I.C2)
		require.Equal(t, clone.I.D2+1111, data.I.D2)
		require.Equal(t, clone.I.E2+1111, data.I.E2)
		require.Equal(t, clone.I.F2+"1111", data.I.F2)
		require.Equal(t, !clone.I.G2, data.I.G2)
	})
}

func TestScanStructENV(t *testing.T) {
	t.Parallel()
	type validSubStruct struct {
		A1 int     `env:"A1"`
		B1 int64   `env:"B1"`
		C1 uint    `env:"C1"`
		D1 uint64  `env:"D1"`
		E1 float64 `env:"E1"`
		F1 string  `env:"F1"`
		G1 bool    `env:"G1"`
	}
	type validSubStruct1 struct {
		A3 int     `env:"A3"`
		B3 int64   `env:"B3"`
		C3 uint    `env:"C3"`
		D3 uint64  `env:"D3"`
		E3 float64 `env:"E3"`
		F3 string  `env:"F3"`
		G3 bool    `env:"G3"`
	}
	type validStruct struct {
		A int     `env:"A"`
		B int64   `env:"B"`
		C uint    `env:"C"`
		D uint64  `env:"D"`
		E float64 `env:"E"`
		F string  `env:"F"`
		G bool    `env:"G"`
		H *validSubStruct
		I validSubStruct1
	}
	type invalidSubStruct struct {
		A2 int `env:"A2"`
	}
	t.Run("error_sub_struct_pointer", func(t *testing.T) {
		t.Parallel()
		type invalidSubStruct struct {
			A2 int `env:"ABCA1"`
		}
		type invalid struct {
			A int `env:"IA"`
			B *invalidSubStruct
		}
		data := invalid{
			A: 1,
			B: &invalidSubStruct{
				A2: 2,
			},
		}
		os.Setenv("ABCA1", "invalid")
		rv := reflect.ValueOf(&data)
		err := env.scanStructENV(rv.Elem())
		require.Error(t, err)
		require.Equal(t, 1, data.A)
		require.Equal(t, 2, data.B.A2)
	})
	t.Run("error_sub_struct", func(t *testing.T) {
		t.Parallel()
		type invalidSubStruct struct {
			A2 int `env:"ABCA2"`
		}
		type invalid struct {
			A int `env:"I1A"`
			B invalidSubStruct
		}
		data := invalid{
			A: 1,
			B: invalidSubStruct{
				A2: 2,
			},
		}
		os.Setenv("ABCA2", "invalid")
		rv := reflect.ValueOf(&data)
		err := env.scanStructENV(rv.Elem())
		require.Error(t, err)
		require.Equal(t, 1, data.A)
		require.Equal(t, 2, data.B.A2)
	})
	t.Run("error_struct", func(t *testing.T) {
		t.Parallel()
		data := struct {
			A2 int `env:"ABCA3"`
		}{
			A2: 1,
		}
		os.Setenv("ABCA3", "invalid")
		rv := reflect.ValueOf(&data)
		err := env.scanStructENV(rv.Elem())
		require.Error(t, err)
		require.Equal(t, 1, data.A2)
	})
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		data := validStruct{
			A: -1,
			B: -10,
			C: 1,
			D: 1000,
			E: 1.8,
			F: "test",
			G: true,
			H: &validSubStruct{
				A1: -2,
				B1: -20,
				C1: 2,
				D1: 2000,
				E1: 2.8,
				F1: "2test",
				G1: true,
			},
			I: validSubStruct1{
				A3: -3,
				B3: -30,
				C3: 3,
				D3: 3000,
				E3: 3.8,
				F3: "3test",
				G3: true,
			},
		}
		clone := validStruct{
			A: -1,
			B: -10,
			C: 1,
			D: 1000,
			E: 1.8,
			F: "test",
			G: true,
			H: &validSubStruct{
				A1: -2,
				B1: -20,
				C1: 2,
				D1: 2000,
				E1: 2.8,
				F1: "2test",
				G1: true,
			},
			I: validSubStruct1{
				A3: -3,
				B3: -30,
				C3: 3,
				D3: 3000,
				E3: 3.8,
				F3: "3test",
				G3: true,
			},
		}
		os.Setenv("A", lib.ToString(data.A+1111))
		os.Setenv("B", lib.ToString(data.B+1111))
		os.Setenv("C", lib.ToString(data.C+1111))
		os.Setenv("D", lib.ToString(data.D+1111))
		os.Setenv("E", lib.ToString(data.E+1111))
		os.Setenv("F", data.F+lib.ToString(1111))
		os.Setenv("G", lib.ToString(!data.G))
		os.Setenv("A1", lib.ToString(data.H.A1+1111))
		os.Setenv("B1", lib.ToString(data.H.B1+1111))
		os.Setenv("C1", lib.ToString(data.H.C1+1111))
		os.Setenv("D1", lib.ToString(data.H.D1+1111))
		os.Setenv("E1", lib.ToString(data.H.E1+1111))
		os.Setenv("F1", data.H.F1+lib.ToString(1111))
		os.Setenv("G1", lib.ToString(!data.H.G1))
		os.Setenv("A3", lib.ToString(data.I.A3+1111))
		os.Setenv("B3", lib.ToString(data.I.B3+1111))
		os.Setenv("C3", lib.ToString(data.I.C3+1111))
		os.Setenv("D3", lib.ToString(data.I.D3+1111))
		os.Setenv("E3", lib.ToString(data.I.E3+1111))
		os.Setenv("F3", data.I.F3+lib.ToString(1111))
		os.Setenv("G3", lib.ToString(!data.I.G3))
		rv := reflect.ValueOf(&data)
		err := env.scanStructENV(rv.Elem())
		require.Nil(t, err)
		require.Equal(t, clone.A+1111, data.A)
		require.Equal(t, clone.B+1111, data.B)
		require.Equal(t, clone.C+1111, data.C)
		require.Equal(t, clone.D+1111, data.D)
		require.Equal(t, clone.E+1111, data.E)
		require.Equal(t, clone.F+"1111", data.F)
		require.Equal(t, !clone.G, data.G)
		require.Equal(t, clone.H.A1+1111, data.H.A1)
		require.Equal(t, clone.H.B1+1111, data.H.B1)
		require.Equal(t, clone.H.C1+1111, data.H.C1)
		require.Equal(t, clone.H.D1+1111, data.H.D1)
		require.Equal(t, clone.H.E1+1111, data.H.E1)
		require.Equal(t, clone.H.F1+"1111", data.H.F1)
		require.Equal(t, !clone.H.G1, data.H.G1)
		require.Equal(t, clone.I.A3+1111, data.I.A3)
		require.Equal(t, clone.I.B3+1111, data.I.B3)
		require.Equal(t, clone.I.C3+1111, data.I.C3)
		require.Equal(t, clone.I.D3+1111, data.I.D3)
		require.Equal(t, clone.I.E3+1111, data.I.E3)
		require.Equal(t, clone.I.F3+"1111", data.I.F3)
		require.Equal(t, !clone.I.G3, data.I.G3)
	})
}

func TestGetFieldENV(t *testing.T) {
	t.Parallel()
	prefix := "test_get_field_env"
	type test struct {
		A int
		B int64
		C uint
		D uint64
		E float64
		F string
		G bool
	}
	data := test{
		A: -1,
		B: -10,
		C: 1,
		D: 1000,
		E: 1.8,
		F: "test",
		G: true,
	}
	t.Run("empty_key_name", func(t *testing.T) {
		t.Parallel()
		clone := data
		rv := reflect.ValueOf(&clone)
		err := env.getFieldENV("A", rv.Elem().FieldByName("A"), "")
		require.Nil(t, err)
		require.Equal(t, -1, clone.A)
	})
	t.Run("notfound_key", func(t *testing.T) {
		t.Parallel()
		clone := data
		rv := reflect.ValueOf(&clone)
		err := env.getFieldENV("A", rv.Elem().FieldByName("A"), prefix+t.Name())
		require.Nil(t, err)
		require.Equal(t, -1, clone.A)
	})
	t.Run("field_cannot_set", func(t *testing.T) {
		t.Parallel()
		clone := data
		rv := reflect.ValueOf(clone)
		os.Setenv(prefix+t.Name(), "1")
		err := env.getFieldENV("A", rv.FieldByName("A"), prefix+t.Name())
		require.Error(t, err)
		require.Equal(t, -1, clone.A)
	})
	testCases := []struct {
		Name      string
		FieldName string
		Value     interface{}
		IsError   bool
	}{
		{"int_parse_env_error", "A", "invalid", true},
		{"int_parse_env_success", "A", data.A + 100, false},
		{"int64_parse_env_error", "B", "invalid", true},
		{"int64_parse_env_success", "B", data.B + 100, false},
		{"uint_parse_env_error", "C", "invalid", true},
		{"uint_parse_env_success", "C", data.C + 100, false},
		{"uint64_parse_env_error", "D", "invalid", true},
		{"uint64_parse_env_success", "D", data.D + 100, false},
		{"float_parse_env_error", "E", "invalid", true},
		{"float_parse_env_success", "E", data.E + 100, false},
		{"string_parse_env_success", "F", data.F + "100", false},
		{"bool_parse_env_error", "G", "invalid", true},
		{"bool_parse_env_success", "G", !data.G, false},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			clone := data
			rv := reflect.ValueOf(&clone)
			key := prefix + tc.Name
			require.Nil(t, os.Setenv(key, lib.ToString(tc.Value)))
			field, ok := rv.Elem().Type().FieldByName(tc.FieldName)
			require.True(t, ok)
			val := rv.Elem().FieldByName(tc.FieldName).Interface()
			err := env.getFieldENV(field.Name, rv.Elem().FieldByName(tc.FieldName), key)
			if tc.IsError {
				require.Error(t, err)
				require.Equal(t, val, rv.Elem().FieldByName(tc.FieldName).Interface())
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.Value, rv.Elem().FieldByName(tc.FieldName).Interface())
			}
		})
	}
}

func TestCreateENV(t *testing.T) {
	t.Parallel()
	t.Run("option error", func(t *testing.T) {
		t.Parallel()
		f1 := func(_ *ENVConfig) error {
			return errors.New("error")
		}
		env, err := CreateENV(f1)
		require.Error(t, err)
		require.Nil(t, env)
	})
	t.Run("success no option", func(t *testing.T) {
		t.Parallel()
		env, err := CreateENV()
		require.Nil(t, err)
		require.NotNil(t, env)
		require.Empty(t, env.Config.Prefix)
	})
	t.Run("success with option", func(t *testing.T) {
		t.Parallel()
		prefix := "TEST_PREFIX"
		env, err := CreateENV(SetPrefixOption(prefix))
		require.Nil(t, err)
		require.NotNil(t, env)
		require.Equal(t, prefix, env.Config.Prefix)
	})
}
