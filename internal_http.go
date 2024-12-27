package crud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func (c Controller) handleHTTPPut(w http.ResponseWriter, r *http.Request, newObjFunc func() interface{}, id string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		c.writeErrText(w, http.StatusInternalServerError, "cannot_read_request_body")
		return
	}

	objClone := newObjFunc()

	if id != "" {
		err2 := c.orm.Load(objClone, id)
		if err2 != nil {
			c.writeErrText(w, http.StatusInternalServerError, "cannot_get_from_db")
			return
		}
		if c.orm.GetObjIDValue(objClone) == 0 {
			c.writeErrText(w, http.StatusNotFound, "not_found_in_db")
			return
		}
	} else {
		c.orm.ResetFields(objClone)
	}

	err = json.Unmarshal(body, objClone)
	if err != nil {
		c.writeErrText(w, http.StatusBadRequest, "invalid_json")
		return
	}

	// Password fields when password function was passed
	if c.passFunc != nil {
		v := reflect.ValueOf(objClone)
		s := v.Elem()
		indir := reflect.Indirect(v)
		typ := indir.Type()
		for j := 0; j < s.NumField(); j++ {
			f := s.Field(j)
			fieldTag := typ.Field(j).Tag.Get(c.tagName)
			gotPassField := false
			if f.Kind() == reflect.String && fieldTag != "" {
				fieldTags := strings.Split(fieldTag, " ")
				for _, ft := range fieldTags {
					if ft == "password" {
						gotPassField = true
						break
					}
				}
			}
			if gotPassField {
				passVal := c.passFunc(f.String())
				s.Field(j).SetString(passVal)
			}
		}
	}

	b, _, err := c.Validate(objClone, nil)
	if !b || err != nil {
		c.writeErrText(w, http.StatusBadRequest, "validation_failed")
		return
	}

	// refactor: fix this by checking if ModifiedAt and ModifiedBy fields exist
	// refactor: also, do the login stuff just like in the ui without umbrella here
	/*opts := stdb.SaveOptions{}
	userId := r.Context().Value(umbrella.UmbrellaContextValue("UmbrellaUserID"))
	rk := reflect.ValueOf(userId).Kind()
	if rk == reflect.Int64 && userId.(int64) != 0 {
		opts.ModifiedBy = userId.(int64)
		opts.ModifiedAt = time.Now().Unix()
	}*/

	err2 := c.orm.Save(objClone)
	if err2 != nil {
		c.writeErrText(w, http.StatusInternalServerError, "cannot_save_to_db")
		return
	}

	if id != "" {
		c.writeOK(w, http.StatusOK, map[string]interface{}{
			"id": c.orm.GetObjIDValue(objClone),
		})
	} else {
		c.writeOK(w, http.StatusCreated, map[string]interface{}{
			"id": c.orm.GetObjIDValue(objClone),
		})
	}
}

func (c Controller) handleHTTPGet(w http.ResponseWriter, r *http.Request, newObjFunc func() interface{}, id string) {
	objClone := newObjFunc()

	hiddenFields := map[string]bool{}
	v := reflect.ValueOf(objClone)
	s := v.Elem()
	indir := reflect.Indirect(v)
	typ := indir.Type()
	for j := 0; j < s.NumField(); j++ {
		f := s.Field(j)
		fieldTag := typ.Field(j).Tag.Get(c.tagName)
		gotHiddenField := false
		if f.Kind() == reflect.String && fieldTag != "" {
			fieldTags := strings.Split(fieldTag, " ")
			for _, ft := range fieldTags {
				if ft == "hidden" {
					gotHiddenField = true
					break
				}
			}
		}
		if gotHiddenField {
			hiddenFields[typ.Field(j).Name] = true
		}
	}

	if id != "" {
		err := c.orm.Load(objClone, id)
		if err != nil {
			c.writeErrText(w, http.StatusInternalServerError, "cannot_get_from_db")
			return
		}

		// hide fields that are tagged with 'hidden'
		for j := 0; j < s.NumField(); j++ {
			f := s.Field(j)
			if f.Kind() == reflect.String && hiddenFields[typ.Field(j).Name] {
				s.Field(j).SetString("(hidden)")
			}
		}

		if c.orm.GetObjIDValue(objClone) == 0 {
			c.writeErrText(w, http.StatusNotFound, "not_found_in_db")
			return
		}

		c.writeOK(w, http.StatusOK, map[string]interface{}{
			"item": objClone,
		})

		return
	}

	// No id, get more elements
	params := c.getParamsFromURI(r.RequestURI)
	limit, _ := strconv.Atoi(params["limit"])
	offset, _ := strconv.Atoi(params["offset"])
	if limit < 1 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	order := []string{}
	if params["order"] != "" {
		order = append(order, params["order"])
		order = append(order, params["order_direction"])
	}

	filters := make(map[string]interface{})
	for k, v := range params {
		if !strings.HasPrefix(k, "filter_") {
			continue
		}
		k = k[7:]
		fieldName, fieldValue, errF := c.uriFilterToFilter(objClone, k, v)
		if errF == nil {
			if fieldName != "" {
				filters[fieldName] = fieldValue
			}
			continue
		}
		if errF.Op == "GetHelper" {
			c.writeErrText(w, http.StatusInternalServerError, "get_helper")
			return
		} else {
			c.writeErrText(w, http.StatusBadRequest, "invalid_filter")
			return
		}
	}

	xobj, err1 := c.orm.Get(newObjFunc, order, limit, offset, filters, func(obj interface{}) interface{} {
		v := reflect.ValueOf(obj)
		s := v.Elem()
		i := reflect.Indirect(v)
		t := i.Type()
		for j := 0; j < s.NumField(); j++ {
			f := s.Field(j)
			if f.Kind() == reflect.String && hiddenFields[t.Field(j).Name] {
				s.Field(j).SetString("(hidden)")
			}
		}
		return obj
	})
	if err1 != nil {
		if err1.IsInvalidFilters() {
			c.writeErrText(w, http.StatusBadRequest, "invalid_filter_value")
			return
		} else {
			c.writeErrText(w, http.StatusInternalServerError, "cannot_get_from_db")
			return
		}
	}

	c.writeOK(w, http.StatusOK, map[string]interface{}{
		"items": xobj,
	})
}

func (c Controller) handleHTTPDelete(w http.ResponseWriter, r *http.Request, newObjFunc func() interface{}, id string) {
	if id == "" {
		c.writeErrText(w, http.StatusBadRequest, "invalid_id")
		return
	}

	objClone := newObjFunc()

	err := c.orm.Load(objClone, id)
	if err != nil {
		c.writeErrText(w, http.StatusInternalServerError, "cannot_get_from_db")
		return
	}
	if c.orm.GetObjIDValue(objClone) == 0 {
		c.writeErrText(w, http.StatusNotFound, "not_found_in_db")
		return
	}

	err = c.orm.Delete(objClone)
	if err != nil {
		c.writeErrText(w, http.StatusInternalServerError, "cannot_delete_from_db")
		return
	}

	c.writeOK(w, http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

func (c Controller) getIDFromURI(uri string, w http.ResponseWriter) (string, bool) {
	xs := strings.SplitN(uri, "?", 2)
	if xs[0] == "" {
		return "", true
	}
	matched, err := regexp.Match(`^[0-9]+$`, []byte(xs[0]))
	if err != nil || !matched {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(c.jsonError("invalid id"))
		return "", false
	}
	return xs[0], true
}

func (c Controller) getParamsFromURI(uri string) map[string]string {
	o := make(map[string]string)
	xs := strings.SplitN(uri, "?", 2)
	if len(xs) < 2 || xs[1] == "" {
		return o
	}
	xp := strings.SplitN(xs[1], "&", -1)
	for _, p := range xp {
		pv := strings.SplitN(p, "=", 2)
		matched, err := regexp.Match(`^[0-9a-zA-Z_]+$`, []byte(pv[0]))
		if len(pv) == 1 || err != nil || !matched {
			continue
		}
		unesc, err := url.QueryUnescape(pv[1])
		if err != nil {
			continue
		}
		o[pv[0]] = unesc
	}
	return o
}

func (c Controller) jsonError(e string) []byte {
	return []byte(fmt.Sprintf("{\"err\":\"%s\"}", e))
}

func (c Controller) jsonID(id int64) []byte {
	return []byte(fmt.Sprintf("{\"id\":\"%d\"}", id))
}

func (c Controller) uriFilterToFilter(obj interface{}, filterName string, filterValue string) (string, interface{}, *ErrController) {
	fieldName, cErr := c.orm.GetFieldNameFromDBCol(obj, filterName)
	if cErr != nil {
		return "", nil, &ErrController{
			Op:  "GetDBCol",
			Err: fmt.Errorf("Error getting field name from filter: %w", cErr),
		}
	}

	if fieldName == "" {
		return "", nil, nil
	}

	val := reflect.ValueOf(obj).Elem()
	valueField := val.FieldByName(fieldName)
	if valueField.Type().Name() == "int" {
		filterInt, err := strconv.Atoi(filterValue)
		if err != nil {
			return "", nil, &ErrController{
				Op:  "InvalidValue",
				Err: fmt.Errorf("Error converting string to int: %w", err),
			}
		}
		return fieldName, filterInt, nil
	}
	if valueField.Type().Name() == "int64" {
		filterInt64, err := strconv.ParseInt(filterValue, 10, 64)
		if err != nil {
			return "", nil, &ErrController{
				Op:  "InvalidValue",
				Err: fmt.Errorf("Error converting string to int64: %w", err),
			}
		}
		return fieldName, filterInt64, nil
	}
	if valueField.Type().Name() == "string" {
		return fieldName, filterValue, nil
	}

	return "", nil, nil
}

func (c Controller) writeErrText(w http.ResponseWriter, status int, errText string) {
	r := NewHTTPResponse(0, errText)
	j, err := json.Marshal(r)
	w.WriteHeader(status)
	if err == nil {
		w.Write(j)
	}
}

func (c Controller) writeOK(w http.ResponseWriter, status int, data map[string]interface{}) {
	r := NewHTTPResponse(1, "")
	r.Data = data
	j, err := json.Marshal(r)
	w.WriteHeader(status)
	if err == nil {
		w.Write(j)
	}
}

func (c *Controller) isStructOperationAllowed(r *http.Request, structName string, op int) bool {
	allowedTypes := r.Context().Value(ContextValue(fmt.Sprintf("AllowedTypes_%d", op)))
	if allowedTypes != nil {
		v, ok := allowedTypes.(map[string]bool)[structName]
		if !ok || !v {
			v2, ok2 := allowedTypes.(map[string]bool)["all"]
			if !ok2 || !v2 {
				return false
			}
		}
	}

	return true
}
