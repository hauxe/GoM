package crudl

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanStructMySQL(t *testing.T) {
	t.Parallel()
	type testScanMySQL struct {
		ID          int64  `json:"id" db:"id_db,pk,select,list"`
		Name        string `json:"name,omitempty" db:"name_db,create,select,update,list,validator=name,optional"`
		Description int64  `json:"description" db:"description_db,create,update,list,validator=des,optional"`
		MissingJSON int64  `db:"missing_json_db,create,select,list"`
		OnlySelect  int64  `json:"only_select" db:"only_select_db,select"`
		OnlyCreate  int64  `json:"only_create,omitempty" db:"only_create_db,create"`
		OnlyUpdate  int64  `json:"only_update,omitempty" db:"only_update_db,update"`
		OnlyList    int64  `json:"only_list,omitempty" db:"only_list_db,list"`
	}
	data := testScanMySQL{}
	rv := reflect.ValueOf(data)
	crud := &CRUD{
		Config: &Config{},
	}
	err := crud.scanStructMySQL(rv)
	require.Nil(t, err)
	require.Equal(t, "id_db", crud.Config.pk)

	createFields := crud.Config.createFields
	require.Len(t, createFields, 4)
	require.Equal(t, "name_db", createFields[0])
	require.Equal(t, "description_db", createFields[1])
	require.Equal(t, "missing_json_db", createFields[2])
	require.Equal(t, "only_create_db", createFields[3])

	updateFields := crud.Config.updateFields
	require.Len(t, updateFields, 3)
	require.Equal(t, "name_db", updateFields[0])
	require.Equal(t, "description_db", updateFields[1])
	require.Equal(t, "only_update_db", updateFields[2])

	selectFields := crud.Config.selectFields
	require.Len(t, selectFields, 4)
	require.Equal(t, "id_db", selectFields[0])
	require.Equal(t, "name_db", selectFields[1])
	require.Equal(t, "missing_json_db", selectFields[2])
	require.Equal(t, "only_select_db", selectFields[3])

	listFields := crud.Config.listFields
	require.Len(t, listFields, 5)
	require.Equal(t, "id_db", listFields[0])
	require.Equal(t, "name_db", listFields[1])
	require.Equal(t, "description_db", listFields[2])
	require.Equal(t, "missing_json_db", listFields[3])
	require.Equal(t, "only_list_db", listFields[4])

}
