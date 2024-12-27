package crud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

var createdID int64

// TestHTTPHanlderPutMethodForValidations checks if HTTP endpoint returns validation failed error when PUT request with invalid input is made
func TestHTTPHandlerPutMethodForValidation(t *testing.T) {
	j := `{
		"email": "invalid",
		"first_name": "J",
		"last_name": "S",
		"key": "12"
	}`
	b := makePUTInsertRequest(j, http.StatusBadRequest, t)
	if !strings.Contains(string(b), "validation_failed") {
		t.Fatalf("PUT method for invalid request did not output validation_failed error text")
	}
}

// TestHTTPHandlerPutMethodForCreating tests if HTTP endpoint properly creates new object in the database, when PUT request is made, without object ID
func TestHTTPHandlerPutMethodForCreating(t *testing.T) {
	j := `{
		"email": "test@example.com",
		"first_name": "John",
		"last_name": "Smith",
		"key": "123456789012345678901234567890aa"
	}`
	b := makePUTInsertRequest(j, http.StatusCreated, t)

	id, flags, primaryEmail, emailSecondary, firstName, lastName, age, price, postCode, postCode2, password, createdByUserID, key, err := getRow()
	if err != nil {
		t.Fatalf("PUT method failed to insert struct to the table: %s", err.Error())
	}
	if id == 0 || flags != 0 || primaryEmail != "test@example.com" || emailSecondary != "" || firstName != "John" || lastName != "Smith" || age != 0 || price != 0 || postCode != "" || postCode2 != "" || createdByUserID != 0 || key != "123456789012345678901234567890aa" || password != "" {
		t.Fatalf("PUT method failed to insert struct to the table")
	}

	r := NewHTTPResponse(1, "")
	err = json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("PUT method returned wrong json output, error marshaling: %s", err.Error())
	}

	if r.Data["id"].(float64) == 0 {
		t.Fatalf("PUT method did not return id")
	}

	createdID = int64(r.Data["id"].(float64))
}

// TestHTTPHandlerPutMethodForUpdating tests if HTTP endpoint successfully updates object details when PUT request with ID is being made
func TestHTTPHandlerPutMethodForUpdating(t *testing.T) {
	j := `{
		"test_struct_flags": 8,
		"email": "test11@example.com",
		"email2": "test22@example.com",
		"first_name": "John2",
		"last_name": "Smith2",
		"age": 39,
		"price": 199,
		"post_code": "22-222",
		"post_code2": "33-333",
		"password": "password123updated",
		"created_by_user_id": 12,
		"key": "123456789012345678901234567890nbh"
	}`
	_ = makePUTUpdateRequest(j, createdID, "", t)

	id, flags, primaryEmail, emailSecondary, firstName, lastName, age, price, postCode, postCode2, password, createdByUserID, key, err := getRow()
	if err != nil {
		t.Fatalf("PUT method failed to update struct to the table: %s", err.Error())
	}
	// Only 2 fields should be updated: FirstName and LastName. Check the TestStruct_Update struct
	if id == 0 || flags != 0 || primaryEmail != "test@example.com" || emailSecondary != "" || firstName != "John2" || lastName != "Smith2" || age != 0 || price != 0 || postCode != "" || postCode2 != "" || createdByUserID != 0 || key != "123456789012345678901234567890aa" || password != "" {
		t.Fatalf("PUT method failed to update struct in the table")
	}
}

// TestHTTPHandlerPutMethodForUpdatingOnCustomEndpoint tests if HTTP endpoint successfully updates object when PUT request with ID
// is being made, and when the endpoint is a custom endpoint (it actually does not matter that much)
func TestHTTPHandlerPutMethodForUpdatingOnCustomEndpoint(t *testing.T) {
	j := `{
		"test_struct_flags": 8,
		"email": "test11@example.com",
		"email2": "test22@example.com",
		"first_name": "John2",
		"last_name": "Smith2",
		"age": 39,
		"price": 444,
		"post_code": "22-222",
		"post_code2": "33-333",
		"password": "password123updated",
		"created_by_user_id": 12,
		"key": "123456789012345678901234567890nbh"
	}`
	_ = makePUTUpdateRequest(j, createdID, httpURI2, t)

	id, flags, _, _, _, _, _, price, _, _, _, _, _, err := getRow()
	if err != nil {
		t.Fatalf("PUT method failed to update struct to the table: %s", err.Error())
	}
	// Only Price field should be updated. Check the TestStruct_Update struct
	if id == 0 || flags != 0 || price != 444 {
		t.Fatalf(strconv.Itoa(price))
		t.Fatalf("PUT method on a custom endpoint failed to update struct in the table")
	}

	j = `{
		"password": "duplicateme!"
	}`
	_ = makePUTUpdateRequest(j, createdID, httpURIPassFunc, t)

	id, flags, _, _, _, _, _, _, _, _, password, _, _, err := getRow()
	if err != nil {
		t.Fatalf("PUT method failed to update struct to the table: %s", err.Error())
	}
	// Only password field should be updated, and its value should be a duplicated string that was passed. Check the TestStruct_UpdatePasswordWithFunc struct
	if id == 0 || flags != 0 || password != "duplicateme!duplicateme!" {
		t.Fatalf("PUT method on a custom endpoint failed to update struct with field using password function in the table (wrong value: %s)", password)
	}

}

// TestHTTPHandlerGetMethodOnExisting checks if HTTP endpoint properly return object details,
// when GET request with object ID is made
func TestHTTPHandlerGetMethodOnExisting(t *testing.T) {
	resp := makeGETReadRequest(createdID, t)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET method returned wrong status code, want %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("GET method failed")
	}

	r := NewHTTPResponse(1, "")
	err = json.Unmarshal(body, &r)
	if err != nil {
		t.Fatalf("GET method failed to return unmarshable JSON: %s", err.Error())
	}
	if r.Data["item"].(map[string]interface{})["age"].(float64) != 0 {
		t.Fatalf("GET method returned invalid values")
	}
	if r.Data["item"].(map[string]interface{})["price"].(float64) != 444 {
		t.Fatalf("GET method returned invalid values")
	}
	if strings.Contains(string(body), "email2") {
		t.Fatalf("GET method returned output with field that should have been hidden")
	}
	if strings.Contains(string(body), "post_code2") {
		t.Fatalf("GET method returned output with field that should have been hidden")
	}
	if r.Data["item"].(map[string]interface{})["password"].(string) != "(hidden)" {
		t.Fatalf("GET method should have hid a field value")
	}
	if r.Data["item"].(map[string]interface{})["first_name"].(string) == "(hidden)" {
		t.Fatalf("GET method hid a field that should not be hidden")
	}
}

// TestHTTPHandlerDeleteMethod tests if HTTP endpoint removes object from the database, when DELETE request is made
func TestHTTPHandlerDeleteMethod(t *testing.T) {
	makeDELETERequest(createdID, t)

	cnt, err2 := getRowCntById(createdID)
	if err2 != nil {
		t.Fatalf("DELETE handler failed to delete struct from the table")
	}
	if cnt > 0 {
		t.Fatalf("DELETE handler failed to delete struct from the table")
	}
}

// TestHTTPHandlerGetMethodOnNonExisting checks HTTP endpoint response when making GET request with non-existing object ID
func TestHTTPHandlerGetMethodOnNonExisting(t *testing.T) {
	resp := makeGETReadRequest(createdID, t)

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET method returned wrong status code, want %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

// TestHTTPHandlerGetMethodWithoutID tests if HTTP endpoint returns list of objects when GET request without ID
// is done; request contains filters, order and result limit
func TestHTTPHandlerGetMethodWithoutID(t *testing.T) {
	truncateTable()

	ts := getTestStructWithData()
	for i := 1; i <= 55; i++ {
		ts.ID = 0
		// Key must be unique
		ts.Key = fmt.Sprintf("%d%s", i, "123456789012345678901234567890")
		ts.Age = ts.Age + 1
		ctl.orm.Save(ts)
	}
	b := makeGETListRequest(map[string]string{
		"limit":                "10",
		"offset":               "20",
		"order":                "age",
		"order_direction":      "asc",
		"filter_price":         "444",
		"filter_primary_email": "primary@example.com",
	}, t)

	r := NewHTTPResponse(1, "")
	err := json.Unmarshal(b, &r)
	if err != nil {
		t.Fatalf("GET method returned wrong json output, error marshaling: %s", err.Error())
	}
	if len(r.Data["items"].([]interface{})) != 10 {
		t.Fatalf("GET method returned invalid number of rows, want %d got %d", 10, len(r.Data["items"].([]interface{})))
	}

	if r.Data["items"].([]interface{})[2].(map[string]interface{})["age"].(float64) != 60 {
		t.Fatalf("GET method returned invalid row, want %d got %f", 60, r.Data["items"].([]interface{})[2].(map[string]interface{})["age"].(float64))
	}

	if strings.Contains(string(b), "email2") {
		t.Fatalf("GET method returned output with field that should have been hidden")
	}
	if strings.Contains(string(b), "post_code2") {
		t.Fatalf("GET method returned output with field that should have been hidden")
	}
	if r.Data["items"].([]interface{})[2].(map[string]interface{})["password"].(string) != "(hidden)" {
		t.Fatalf("GET method without id should have hid a field value")
	}
	if r.Data["items"].([]interface{})[2].(map[string]interface{})["first_name"].(string) == "(hidden)" {
		t.Fatalf("GET method without id hid a field that should not be hidden")
	}
}
