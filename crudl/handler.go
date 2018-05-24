package crudl

import (
	"net/http"

	gomHTTP "github.com/hauxe/GoM/http"
)

type response struct {
	status       int
	errorCode    int
	errorMessage string
	data         map[string]interface{}
}

func (crud *CRUD) handleCreate(w http.ResponseWriter, r *http.Request) {
	obj := crud.Config.Object.Get()

	err := gomHTTP.ParseParameters(r, obj)
	if err != nil {
		err = gomHTTP.SendError(w, err)
		if err != nil {
			crud.Logger.For(r.Context()).Error(err.Error())
		}
		return
	}
	err = crud.Create(obj)
	if err != nil {
		err = gomHTTP.SendError(w, err)
		if err != nil {
			crud.Logger.For(r.Context()).Error(err.Error())
		}
	}

	err = gomHTTP.SendResponse(w, http.StatusOK, gomHTTP.ErrorCodeSuccess, "created successfully", map[string]interface{}{
		"success": obj,
	})
	if err != nil {
		crud.Logger.For(r.Context()).Error(err.Error())
	}
}

func (crud *CRUD) handleRead(w http.ResponseWriter, r *http.Request) {
	obj := map[string]interface{}{}

	err := gomHTTP.ParseParameters(r, obj)
	if err != nil {
		err = gomHTTP.SendError(w, err)
		if err != nil {
			crud.Logger.For(r.Context()).Error(err.Error())
		}
		return
	}
	row, err := crud.Read(obj[crud.Config.pk])
	if err != nil {
		err = gomHTTP.SendError(w, err)
		if err != nil {
			crud.Logger.For(r.Context()).Error(err.Error())
		}
	}

	err = gomHTTP.SendResponse(w, http.StatusOK, gomHTTP.ErrorCodeSuccess, "created successfully", map[string]interface{}{
		"success": row,
	})
	if err != nil {
		crud.Logger.For(r.Context()).Error(err.Error())
	}
}

func (crud *CRUD) handleUpdate(w http.ResponseWriter, r *http.Request) {
	obj := crud.Config.Object.Get()

	err := gomHTTP.ParseParameters(r, obj)
	if err != nil {
		err = gomHTTP.SendError(w, err)
		if err != nil {
			crud.Logger.For(r.Context()).Error(err.Error())
		}
		return
	}
	err = crud.Update(obj)
	if err != nil {
		err = gomHTTP.SendError(w, err)
		if err != nil {
			crud.Logger.For(r.Context()).Error(err.Error())
		}
	}

	err = gomHTTP.SendResponse(w, http.StatusOK, gomHTTP.ErrorCodeSuccess, "updated successfully", map[string]interface{}{
		"success": obj,
	})
	if err != nil {
		crud.Logger.For(r.Context()).Error(err.Error())
	}
}

func (crud *CRUD) handleDelete(w http.ResponseWriter, r *http.Request) {
	obj := map[string]interface{}{}

	err := gomHTTP.ParseParameters(r, obj)
	if err != nil {
		err = gomHTTP.SendError(w, err)
		if err != nil {
			crud.Logger.For(r.Context()).Error(err.Error())
		}
		return
	}
	row, err := crud.Delete(obj[crud.Config.pk])
	if err != nil {
		err = gomHTTP.SendError(w, err)
		if err != nil {
			crud.Logger.For(r.Context()).Error(err.Error())
		}
	}

	err = gomHTTP.SendResponse(w, http.StatusOK, gomHTTP.ErrorCodeSuccess, "created successfully", map[string]interface{}{
		"success": row,
	})
	if err != nil {
		crud.Logger.For(r.Context()).Error(err.Error())
	}
}

// handleList handle request for getting list data
func (crud *CRUD) handleList(w http.ResponseWriter, r *http.Request) {
	obj := struct {
		PageID  int64 `json:"page_id"`
		PerPage int64 `json:"per_page"`
	}{}

	err := gomHTTP.ParseParameters(r, obj)
	if err != nil {
		err = gomHTTP.SendError(w, err)
		if err != nil {
			crud.Logger.For(r.Context()).Error(err.Error())
		}
		return
	}
	l, err := crud.List(obj.PageID, obj.PerPage)
	if err != nil {
		err = gomHTTP.SendError(w, err)
		if err != nil {
			crud.Logger.For(r.Context()).Error(err.Error())
		}
	}

	err = gomHTTP.SendResponse(w, http.StatusOK, gomHTTP.ErrorCodeSuccess, "updated successfully", map[string]interface{}{
		"success": l,
	})
	if err != nil {
		crud.Logger.For(r.Context()).Error(err.Error())
	}
}
