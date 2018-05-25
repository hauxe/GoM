package crudl

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"

	gomHTTP "github.com/hauxe/gom/http"
	lib "github.com/hauxe/gom/library"
	gomMySQL "github.com/hauxe/gom/mysql"
	"github.com/stretchr/testify/require"
)

var sampleValidator = func(method string, obj interface{}) (err error) {
	if method == "create" {
		if c, ok := obj.(*create); ok {
			if c.Name == "" {
				return errors.New("require name")
			}
			if c.Description == "test_failed_validator" {
				return errors.New("sample validation error")
			}
		}
	} else if method == "update" {
		if c, ok := obj.(*update); ok {
			if c.Description == "test_failed_validator" {
				return errors.New("sample validation error")
			}
		}
	}
	return
}

type create struct {
	ID          int64                `json:"id" db:"id,pk"`
	Name        string               `json:"name" db:"name,create"`
	Age         int                  `json:"age" db:"age,create"`
	Description string               `json:"description" db:"description,create,validator=description"`
	Info        gomMySQL.StringMap   `json:"info" db:"info,create"`
	List        gomMySQL.StringSlice `json:"list" db:"list,create"`
	CreatedAt   time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at" db:"updated_at"`
}

type get struct {
	ID          int64                `json:"id" db:"id,pk" schema:"id,required"`
	Name        string               `json:"name" db:"name,create"`
	Age         int                  `json:"age" db:"age,create"`
	Description string               `json:"description" db:"description,create,validator=description"`
	Info        gomMySQL.StringMap   `json:"info" db:"info,create"`
	List        gomMySQL.StringSlice `json:"list" db:"list,create"`
	CreatedAt   time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at" db:"updated_at"`
}

type update struct {
	ID          int64                `json:"id" db:"id,pk"`
	Name        string               `json:"name" db:"name,create"`
	Age         int                  `json:"age" db:"age,create,update"`
	Description string               `json:"description" db:"description,create,update,validator=description"`
	Info        gomMySQL.StringMap   `json:"info" db:"info,create,update"`
	List        gomMySQL.StringSlice `json:"list" db:"list,create,update"`
	CreatedAt   time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at" db:"updated_at"`
}

type delete struct {
	ID          int64                `json:"id" db:"id,pk"`
	Name        string               `json:"name" db:"name,create"`
	Age         int                  `json:"age" db:"age,create"`
	Description string               `json:"description" db:"description,create,validator=description"`
	Info        gomMySQL.StringMap   `json:"info" db:"info,create"`
	List        gomMySQL.StringSlice `json:"list" db:"list,create"`
	CreatedAt   time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at" db:"updated_at"`
}

type list struct {
	ID          int64                `json:"id" db:"id,pk"`
	Name        string               `json:"name" db:"name,create"`
	Age         int                  `json:"age" db:"age,create"`
	Description string               `json:"description" db:"description,create,validator=description"`
	Info        gomMySQL.StringMap   `json:"info" db:"info,create"`
	List        gomMySQL.StringSlice `json:"list" db:"list,create"`
	CreatedAt   time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at" db:"updated_at"`
}

type createCRUD struct{}

func (c *createCRUD) Get() interface{} {
	return &create{}
}

type getCRUD struct{}

func (c *getCRUD) Get() interface{} {
	return &get{}
}

type updateCRUD struct{}

func (c *updateCRUD) Get() interface{} {
	return &update{}
}

type deleteCRUD struct{}

func (c *deleteCRUD) Get() interface{} {
	return &delete{}
}

type listCRUD struct{}

func (c *listCRUD) Get() interface{} {
	return &list{}
}

func TestCRUDCreateHandler(t *testing.T) {
	t.Parallel()
	dropTableSQL := "DROP TABLE IF EXISTS test_crud_create_handler"
	createTableSQL := `CREATE TABLE IF NOT EXISTS test_crud_create_handler (
		id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
		name VARCHAR(31) NOT NULL,
		age INT(11) NOT NULL DEFAULT 0,
		description VARCHAR(255) NOT NULL DEFAULT '',
		info JSON,
		list VARCHAR(255) NOT NULL DEFAULT '[]',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id))
	  ENGINE = InnoDB;`
	_, err := sampleDB.Exec(dropTableSQL)
	require.Nil(t, err)
	_, err = sampleDB.Exec(createTableSQL)
	require.Nil(t, err)

	crud, routes, err := Register(sampleDB, "test_crud_create_handler", &createCRUD{}, UseC(),
		SetValidators(map[string]Validator{
			"description": sampleValidator,
		}))
	require.Nil(t, err)
	require.Len(t, routes, 1)
	require.NotNil(t, crud)
	server, err := CreateSampleServer(routes...)
	require.Nil(t, err)
	require.NotNil(t, server)

	validData := &create{
		Name:        "test_create_handler",
		Age:         11,
		Description: "test_create_handle_description",
		Info:        gomMySQL.StringMap{"a": "a", "b": "b", "c": "c"},
		List:        gomMySQL.StringSlice{"a", "b", "c"},
	}
	inValidData := map[string]interface{}{
		"name":        "invalid \n name",
		"age":         "invalid",
		"description": "invalid \n description",
		"info":        "invalid",
		"list":        "invalid",
	}
	validNoOptionalData := &create{
		Name: "create_handler_no_optional",
	}
	missingRequiredParamData := &create{
		Age:         11,
		Description: "description",
		Info:        gomMySQL.StringMap{"a": "a", "b": "b", "c": "c"},
		List:        gomMySQL.StringSlice{"a", "b", "c"},
	}
	failedValidatorData := &create{
		Name:        "test_create_handler",
		Age:         11,
		Description: "test_failed_validator",
		Info:        gomMySQL.StringMap{"a": "a", "b": "b", "c": "c"},
		List:        gomMySQL.StringSlice{"a", "b", "c"},
	}

	testCases := []struct {
		Name       string
		Data       interface{}
		StatusCode int
		ErrorCode  gomHTTP.ErrorCode
	}{
		{"invalid", inValidData, http.StatusBadRequest, gomHTTP.ErrorCodeBadRequest},
		{"missing_required_param", missingRequiredParamData, http.StatusBadRequest, gomHTTP.ErrorCodeValidationFailed},
		{"failed_validation", failedValidatorData, http.StatusBadRequest, gomHTTP.ErrorCodeValidationFailed},
		{"success", validData, http.StatusOK, gomHTTP.ErrorCodeSuccess},
		{"success_no_optional", validNoOptionalData, http.StatusOK, gomHTTP.ErrorCodeSuccess},
	}
	client, err := gomHTTP.CreateClient()
	require.Nil(t, err)
	require.NotNil(t, client)
	client.Connect()

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			resp, err := client.Send(context.Background(), routes[0].Method, server.URL+routes[0].Path,
				client.SetRequestOptionJSON(tc.Data))
			require.Nil(t, err)
			data := &create{}
			response := gomHTTP.ServerResponse{
				Data: gomHTTP.ServerResponseData{
					Success: data,
				},
			}
			require.Equal(t, tc.StatusCode, resp.StatusCode)
			err = client.ParseJSON(resp, &response)
			require.Nil(t, err)
			require.Equal(t, tc.ErrorCode, response.ErrorCode)
			if tc.ErrorCode == gomHTTP.ErrorCodeSuccess {
				c, ok := tc.Data.(*create)
				require.True(t, ok)
				assertCreate(t, c, data)
			}
		})
	}
}

func TestCRUDGetHandler(t *testing.T) {
	t.Parallel()
	dropTableSQL := "DROP TABLE IF EXISTS test_crud_get_handler"
	createTableSQL := `CREATE TABLE IF NOT EXISTS test_crud_get_handler (
		id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
		name VARCHAR(31) NOT NULL,
		age INT(11) NOT NULL DEFAULT 0,
		description VARCHAR(255) NOT NULL DEFAULT '',
		info JSON,
		list VARCHAR(255) NOT NULL DEFAULT '[]',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id))
	  ENGINE = InnoDB;`
	_, err := sampleDB.Exec(dropTableSQL)
	require.Nil(t, err)
	_, err = sampleDB.Exec(createTableSQL)
	require.Nil(t, err)

	crud, routes, err := Register(sampleDB, "test_crud_get_handler", &getCRUD{}, UseC(), UseR())
	require.Nil(t, err)
	require.Len(t, routes, 2)
	require.NotNil(t, crud)
	server, err := CreateSampleServer(routes...)
	require.Nil(t, err)
	require.NotNil(t, server)
	testData := &get{
		Name:        "test_name",
		Age:         10,
		Description: "this is a description",
		Info:        gomMySQL.StringMap{"a": "a", "b": "b", "c": "c"},
		List:        gomMySQL.StringSlice{"a", "b", "c"},
	}
	err = crud.Create(testData)
	require.Nil(t, err)
	require.True(t, testData.ID > 0)

	validData := map[string]interface{}{
		crud.Config.pk.name: testData.ID,
	}
	inValidData := map[string]interface{}{
		crud.Config.pk.name: "invalid",
	}

	testCases := []struct {
		Name       string
		Data       map[string]interface{}
		StatusCode int
		ErrorCode  gomHTTP.ErrorCode
	}{
		{"invalid", inValidData, http.StatusBadRequest, gomHTTP.ErrorCodeBadRequest},
		{"missing_param", map[string]interface{}{}, http.StatusBadRequest, gomHTTP.ErrorCodeBadRequest},
		{"success", validData, http.StatusOK, gomHTTP.ErrorCodeSuccess},
	}
	client, err := gomHTTP.CreateClient()
	require.Nil(t, err)
	require.NotNil(t, client)
	client.Connect()

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			resp, err := client.Send(context.Background(), routes[1].Method, server.URL+routes[1].Path,
				client.SetRequestOptionQuery(tc.Data))
			require.Nil(t, err)
			data := &get{}
			response := gomHTTP.ServerResponse{
				Data: gomHTTP.ServerResponseData{
					Success: data,
				},
			}
			require.Equal(t, tc.StatusCode, resp.StatusCode)
			err = client.ParseJSON(resp, &response)
			require.Nil(t, err)
			require.Equal(t, tc.ErrorCode, response.ErrorCode)
			if tc.ErrorCode == gomHTTP.ErrorCodeSuccess {
				assertGet(t, testData, data)
			}
		})
	}
}

func TestCRUDUpdateHandler(t *testing.T) {
	t.Parallel()
	dropTableSQL := "DROP TABLE IF EXISTS test_crud_update_handler"
	createTableSQL := `CREATE TABLE IF NOT EXISTS test_crud_update_handler (
		id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
		name VARCHAR(31) NOT NULL DEFAULT '',
		age INT(11) NOT NULL DEFAULT 0,
		description VARCHAR(255) NOT NULL DEFAULT '',
		info JSON,
		list VARCHAR(255) NOT NULL DEFAULT '[]',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id))
	  ENGINE = InnoDB;`
	_, err := sampleDB.Exec(dropTableSQL)
	require.Nil(t, err)
	_, err = sampleDB.Exec(createTableSQL)
	require.Nil(t, err)
	crud, routes, err := Register(sampleDB, "test_crud_update_handler", &updateCRUD{}, UseC(), UseU(),
		SetValidators(map[string]Validator{
			"description": sampleValidator,
		}))
	require.Nil(t, err)
	require.Len(t, routes, 2)
	require.NotNil(t, crud)
	server, err := CreateSampleServer(routes...)
	require.Nil(t, err)
	require.NotNil(t, server)
	testData := &update{
		Name:        "test_name",
		Age:         10,
		Description: "this is a description",
		Info:        gomMySQL.StringMap{"a": "a", "b": "b", "c": "c"},
		List:        gomMySQL.StringSlice{"a", "b", "c"},
	}
	err = crud.Create(testData)
	require.Nil(t, err)
	require.True(t, testData.ID > 0)

	validData := &update{
		ID:          testData.ID,
		Age:         20,
		Description: "test_updated_description",
		Info:        gomMySQL.StringMap{"a": "a1", "b": "b1", "c": "c1"},
		List:        gomMySQL.StringSlice{"a1", "b1", "c1"},
	}
	inValidData := map[string]interface{}{
		crud.Config.pk.name: "invalid",
		"age":               "invalid",
		"description":       "invalid \n description",
		"info":              "invalid",
		"list":              "invalid",
	}
	failedValidatorData := &update{
		ID:          testData.ID,
		Age:         20,
		Description: "test_failed_validator",
		Info:        gomMySQL.StringMap{"a": "a1", "b": "b1", "c": "c1"},
		List:        gomMySQL.StringSlice{"a1", "b1", "c1"},
	}

	testCases := []struct {
		Name       string
		Data       interface{}
		StatusCode int
		ErrorCode  gomHTTP.ErrorCode
	}{
		{"invalid", inValidData, http.StatusBadRequest, gomHTTP.ErrorCodeBadRequest},
		{"missing_param", map[string]interface{}{}, http.StatusBadRequest, gomHTTP.ErrorCodeValidationFailed},
		{"failed_validation", failedValidatorData, http.StatusBadRequest, gomHTTP.ErrorCodeValidationFailed},
		{"success", validData, http.StatusOK, gomHTTP.ErrorCodeSuccess},
	}

	client, err := gomHTTP.CreateClient()
	require.Nil(t, err)
	require.NotNil(t, client)
	client.Connect()

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			resp, err := client.Send(context.Background(), routes[1].Method, server.URL+routes[1].Path,
				client.SetRequestOptionJSON(tc.Data))
			require.Nil(t, err)
			data := &update{}
			response := gomHTTP.ServerResponse{
				Data: gomHTTP.ServerResponseData{
					Success: data,
				},
			}
			require.Equal(t, tc.StatusCode, resp.StatusCode)
			err = client.ParseJSON(resp, &response)
			require.Nil(t, err)
			require.Equal(t, tc.ErrorCode, response.ErrorCode)
			if tc.ErrorCode == gomHTTP.ErrorCodeSuccess {
				c, ok := tc.Data.(*update)
				require.True(t, ok)
				assertUpdate(t, c, data)
			}
		})
	}
}

func TestCRUDDeleteHandler(t *testing.T) {
	t.Parallel()
	dropTableSQL := "DROP TABLE IF EXISTS test_crud_delete_handler"
	createTableSQL := `CREATE TABLE IF NOT EXISTS test_crud_delete_handler (
		id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
		name VARCHAR(31) NOT NULL DEFAULT '',
		age INT(11) NOT NULL DEFAULT 0,
		description VARCHAR(255) NOT NULL DEFAULT '',
		info JSON,
		list VARCHAR(255) NOT NULL DEFAULT '[]',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id))
	  ENGINE = InnoDB;`
	_, err := sampleDB.Exec(dropTableSQL)
	require.Nil(t, err)
	_, err = sampleDB.Exec(createTableSQL)
	require.Nil(t, err)

	crud, routes, err := Register(sampleDB, "test_crud_delete_handler", &deleteCRUD{}, UseC(), UseD(), UseR())
	require.Nil(t, err)
	require.Len(t, routes, 3)
	require.NotNil(t, crud)
	server, err := CreateSampleServer(routes...)
	require.Nil(t, err)
	require.NotNil(t, server)
	testData := &delete{
		Name:        "test_name",
		Age:         10,
		Description: "this is a description",
		Info:        gomMySQL.StringMap{"a": "a", "b": "b", "c": "c"},
		List:        gomMySQL.StringSlice{"a", "b", "c"},
	}
	err = crud.Create(testData)
	require.Nil(t, err)
	require.True(t, testData.ID > 0)

	validData := &delete{
		ID: testData.ID,
	}
	inValidData := map[string]interface{}{
		crud.Config.pk.name: "invalid",
	}

	testCases := []struct {
		Name       string
		Data       interface{}
		StatusCode int
		ErrorCode  gomHTTP.ErrorCode
	}{
		{"invalid", inValidData, http.StatusBadRequest, gomHTTP.ErrorCodeBadRequest},
		{"missing_param", map[string]interface{}{}, http.StatusBadRequest, gomHTTP.ErrorCodeValidationFailed},
		{"success", validData, http.StatusOK, gomHTTP.ErrorCodeSuccess},
	}

	client, err := gomHTTP.CreateClient()
	require.Nil(t, err)
	require.NotNil(t, client)
	client.Connect()

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			resp, err := client.Send(context.Background(), routes[1].Method, server.URL+routes[1].Path,
				client.SetRequestOptionJSON(tc.Data))
			require.Nil(t, err)
			require.Equal(t, tc.StatusCode, resp.StatusCode)
			if tc.ErrorCode == gomHTTP.ErrorCodeSuccess {
				obj, err := crud.Read(testData)
				require.Nil(t, err)
				require.Nil(t, obj)
			}
		})
	}
}

func TestCRUDListHandler(t *testing.T) {
	t.Parallel()
	dropTableSQL := "DROP TABLE IF EXISTS test_crud_list_handler"
	createTableSQL := `CREATE TABLE IF NOT EXISTS test_crud_list_handler (
		id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
		name VARCHAR(31) NOT NULL DEFAULT '',
		age INT(11) NOT NULL DEFAULT 0,
		description VARCHAR(255) NOT NULL DEFAULT '',
		info JSON,
		list VARCHAR(255) NOT NULL DEFAULT '[]',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id))
	  ENGINE = InnoDB;`
	_, err := sampleDB.Exec(dropTableSQL)
	require.Nil(t, err)
	_, err = sampleDB.Exec(createTableSQL)
	require.Nil(t, err)

	crud, routes, err := Register(sampleDB, "test_crud_list_handler", &listCRUD{}, UseC(), UseL())
	require.Nil(t, err)
	require.Len(t, routes, 2)
	require.NotNil(t, crud)
	server, err := CreateSampleServer(routes...)
	require.Nil(t, err)
	require.NotNil(t, server)
	n := 10
	testData := make([]*list, n)
	for i := 0; i < n; i++ {
		testData[i] = &list{
			Name:        "test_list_name" + lib.ToString(i),
			Age:         10 + i,
			Description: "this is a test list description" + lib.ToString(i),
			Info:        gomMySQL.StringMap{"a": "a", "b": "b", "c": "c"},
			List:        gomMySQL.StringSlice{"a", "b", "c"},
		}
		// create data first
		err := crud.Create(testData[i])
		require.Nil(t, err)
		require.True(t, testData[i].ID > 0)
	}

	validData := map[string]interface{}{
		"page_id":  int64(1),
		"per_page": int64(n),
	}
	inValidData := map[string]interface{}{
		"page_id":  "invalid",
		"per_page": "invalid",
	}
	testCases := []struct {
		Name       string
		Data       map[string]interface{}
		StatusCode int
		ErrorCode  gomHTTP.ErrorCode
	}{
		{"invalid", inValidData, http.StatusBadRequest, gomHTTP.ErrorCodeBadRequest},
		{"missing_param", map[string]interface{}{}, http.StatusBadRequest, gomHTTP.ErrorCodeBadRequest},
		{"success", validData, http.StatusOK, gomHTTP.ErrorCodeSuccess},
	}

	client, err := gomHTTP.CreateClient()
	require.Nil(t, err)
	require.NotNil(t, client)
	client.Connect()

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			resp, err := client.Send(context.Background(), routes[1].Method, server.URL+routes[1].Path,
				client.SetRequestOptionQuery(tc.Data))
			require.Nil(t, err)
			data := []*list{}
			response := gomHTTP.ServerResponse{
				Data: gomHTTP.ServerResponseData{
					Success: &data,
				},
			}
			require.Equal(t, tc.StatusCode, resp.StatusCode)
			err = client.ParseJSON(resp, &response)
			require.Nil(t, err)
			require.Equal(t, tc.ErrorCode, response.ErrorCode)
			if tc.ErrorCode == gomHTTP.ErrorCodeSuccess {
				require.Len(t, data, n)
				for i, d := range data {
					assertList(t, testData[n-i-1], d)
				}
			}
		})
	}
}

func assertCreate(t *testing.T, expected *create, actual *create) {
	require.NotNil(t, expected)
	require.NotNil(t, actual)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Age, actual.Age)
	require.Equal(t, expected.Description, actual.Description)
	require.True(t, reflect.DeepEqual(expected.Info, actual.Info))
	require.ElementsMatch(t, expected.List, actual.List)
	require.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Second)
	require.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt, time.Second)
}

func assertGet(t *testing.T, expected *get, actual *get) {
	require.NotNil(t, expected)
	require.NotNil(t, actual)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Age, actual.Age)
	require.Equal(t, expected.Description, actual.Description)
	require.True(t, reflect.DeepEqual(expected.Info, actual.Info))
	require.ElementsMatch(t, expected.List, actual.List)
	require.False(t, actual.CreatedAt.IsZero())
	require.False(t, actual.UpdatedAt.IsZero())
}

func assertUpdate(t *testing.T, expected *update, actual *update) {
	require.NotNil(t, expected)
	require.NotNil(t, actual)
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.Age, actual.Age)
	require.Equal(t, expected.Description, actual.Description)
	require.True(t, reflect.DeepEqual(expected.Info, actual.Info))
	require.ElementsMatch(t, expected.List, actual.List)
}

func assertList(t *testing.T, expected *list, actual *list) {
	require.NotNil(t, expected)
	require.NotNil(t, actual)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Age, actual.Age)
	require.Equal(t, expected.Description, actual.Description)
	require.True(t, reflect.DeepEqual(expected.Info, actual.Info))
	require.ElementsMatch(t, expected.List, actual.List)
	require.False(t, actual.CreatedAt.IsZero())
	require.False(t, actual.UpdatedAt.IsZero())
}
