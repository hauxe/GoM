package crudl

import (
	"reflect"
	"strings"

	sdklog "github.com/hauxe/gom/log"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	sqlTag       = "db"
	sqlPK        = "pk"
	sqlValidator = "validator" //format: validator=[name]
	sqlCreate    = "create"
	sqlUpdate    = "update"
	sqlSelect    = "select" // empty means select all
	sqlList      = "list"   // empty means select all
)

const (
	sqlCRUDCreate = "INSERT INTO `%s` (%s) VALUES (%s)"
	sqlCRUDRead   = "SELECT %s FROM `%s` WHERE %s = ?"
	sqlCRUDUpdate = "UPDATE `%s` SET %s WHERE %s = :%s;"
	sqlCRUDDelete = "DELETE FROM `%s` WHERE %s = ?;"
	sqlCRUDList   = "SELECT %s FROM `%s` ORDER by %s DESC LIMIT ?,?;"
)

// Option defines option functions type
type Option func(*Config) error

// Object target crud object
type Object interface {
	Get() interface{}
}

// Validator defines crud validator
type Validator func(method string, obj interface{}) error

type crud int

// CRUD constant type
const (
	CREATE crud = iota
	READ
	UPDATE
	DELETE
)

// Config defines crud properties
type Config struct {
	DB              *sqlx.DB
	TableName       string
	Object          Object
	L               bool
	C               bool
	R               bool
	U               bool
	D               bool
	Validators      map[string]Validator
	fields          []string
	createFields    []string
	updateFields    []string
	selectFields    []string
	listFields      []string
	pk              string
	fieldValidators map[string]string
	sqlCRUDCreate   string
	sqlCRUDRead     string
	sqlCRUDUpdate   string
	sqlCRUDDelete   string
	sqlCRUDList     string
}

// CRUD defines crud properties
type CRUD struct {
	Config *Config
	Logger *sdklog.Factory
}

// scanStructMySQL scan sql row by struct
func (crud *CRUD) scanStructMySQL(rv reflect.Value) (err error) {
	for i := 0; i < rv.NumField(); i++ {
		tag, ok := rv.Type().Field(i).Tag.Lookup(sqlTag)
		if !ok {
			continue
		}
		tags := strings.Split(tag, ",")
		n := len(tags)
		// was this field marked for skipping?
		if tags[0] == "-" {
			continue
		}
		if rv.Field(i).Kind() == reflect.Ptr && rv.Field(i).Elem().Kind() == reflect.Struct {
			continue
		}
		field := tags[0]
		if field == "" {
			// search for tag name in json instead
			jsonTag, ok := rv.Type().Field(i).Tag.Lookup("json")
			if ok && jsonTag != "" {
				jsons := strings.Split(jsonTag, ",")
				field = jsons[0]
			}
		}
		if field == "" {
			// skip this field
			continue
		}
		crud.Config.fields = append(crud.Config.fields, field)
		if crud.Config.fieldValidators == nil {
			crud.Config.fieldValidators = make(map[string]string)
		}
		for j := 1; j < n; j++ {
			switch tags[j] {
			case sqlCreate:
				crud.Config.createFields = append(crud.Config.createFields, field)
			case sqlUpdate:
				crud.Config.updateFields = append(crud.Config.updateFields, field)
			case sqlSelect:
				crud.Config.selectFields = append(crud.Config.selectFields, field)
			case sqlList:
				crud.Config.listFields = append(crud.Config.listFields, field)
			case sqlPK:
				crud.Config.pk = field
			default:
				vals := strings.Split(tags[j], "=")
				if len(vals) == 2 {
					switch vals[0] {
					case sqlValidator:
						crud.Config.fieldValidators[field] = vals[1]
					}
				}
			}
		}
	}
	return
}

// Create creates from map
func (crud *CRUD) Create(data interface{}) error {
	// build fields
	result, err := crud.Config.DB.NamedExec(crud.Config.sqlCRUDCreate, data)
	if err != nil {
		return errors.Wrap(err, "error crud create")
	}
	if crud.Config.pk != "" {
		// set primary key
		rv := reflect.ValueOf(data)
		rv = reflect.Indirect(rv)
		pk := rv.FieldByName(crud.Config.pk)
		if pk.CanSet() {
			id, err := result.LastInsertId()
			if err != nil {
				return errors.Wrap(err, "error get last insert id at crud create")
			}
			pk.SetInt(id)
		}
	}
	return nil
}

// Read read data
func (crud *CRUD) Read(pk interface{}) (interface{}, error) {
	rows, err := crud.Config.DB.Queryx(crud.Config.sqlCRUDRead, pk)
	if err != nil {
		return nil, errors.Wrap(err, "error crud read")
	}
	defer rows.Close()
	if rows.Next() {
		obj := crud.Config.Object.Get()
		err = rows.StructScan(obj)
		if err != nil {
			return nil, errors.Wrap(err, "error crud scan")
		}
		return obj, nil
	}
	return nil, nil
}

// Update update data
func (crud *CRUD) Update(data interface{}) error {
	_, err := crud.Config.DB.NamedExec(crud.Config.sqlCRUDUpdate, data)
	if err != nil {
		return errors.Wrap(err, "error crud update")
	}
	return nil
}

// Delete delete row
func (crud *CRUD) Delete(pk interface{}) (int64, error) {
	result, err := crud.Config.DB.Exec(crud.Config.sqlCRUDDelete, pk)
	if err != nil {
		return 0, errors.Wrap(err, "error crud delete")
	}
	return result.RowsAffected()
}

// List lists data and paging the result
func (crud *CRUD) List(pageID, perPage int64) ([]interface{}, error) {
	offset := (pageID - 1) * perPage
	rows, err := crud.Config.DB.Queryx(crud.Config.sqlCRUDList, offset, perPage)
	if err != nil {
		return nil, errors.Wrap(err, "error crud list")
	}
	defer rows.Close()

	result := []interface{}{}
	for rows.Next() {
		obj := crud.Config.Object.Get()
		err = rows.StructScan(obj)
		if err != nil {
			return nil, errors.Wrap(err, "error crud scan")
		}
		result = append(result, obj)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "error crud loop rows list")
	}
	return result, nil
}
