package crud

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

type ProductKind struct {
	ID   int64
	Name string
}

type ProductGroup struct {
	ID          int64
	Name        string
	Description string
	Code        string
}

type Product struct {
	ID            int64
	Name          string
	Price         int
	ProductKindID int64
	ProductGrpID  int64
}

type Product_WithDetails struct {
	ID               int64         `json:"product_id"`
	Name             string        `json:"name"`
	Price            int           `json:"price"`
	ProductKindID    int64         `json:"product_kind_id"`
	ProductGrpID     int64         `json:"product_grp_id"`
	ProductKind      *ProductKind  `crud:"join" json:"omit"`
	ProductKind_Name string        `json:"product_kind_name"`
	ProductGrp       *ProductGroup `crud:"join" json:"omit"`
	ProductGrp_Code  string        `json:"product_grp_code"`
}

func TestGetListOfJoinedStructs(t *testing.T) {
	// Create tables
	ctl.orm.CreateTables(&Product{}, &ProductGroup{}, &ProductKind{})

	// Add rows
	pg := &ProductGroup{
		ID:          113,
		Name:        "Group 1",
		Description: "A group of products",
		Code:        "GRP1",
	}
	ctl.orm.Save(pg)

	pk := &ProductKind{
		ID:   33,
		Name: "Kind 1",
	}
	ctl.orm.Save(pk)
	pk2 := &ProductKind{
		ID:   34,
		Name: "Kind 2",
	}
	ctl.orm.Save(pk2)

	p := &Product{
		ID:            6,
		Name:          "Product Name",
		Price:         1234,
		ProductKindID: 33,
		ProductGrpID:  113,
	}
	ctl.orm.Save(p)
	p2 := &Product{
		ID:            7,
		Name:          "Product Name 2",
		Price:         1234,
		ProductKindID: 34,
		ProductGrpID:  113,
	}
	ctl.orm.Save(p2)

	uriParamString := ""
	req, err := http.NewRequest("GET", "http://localhost:"+httpPort+httpURIJoined+"?"+uriParamString, bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatalf("GET method failed on HTTP server with handler from GetHTTPHandler: %s", err)
	}
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("GET method failed on HTTP server with handler from GetHTTPHandler: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET method returned wrong status code, want %d, got %d", http.StatusOK, resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("GET method failed to return body: %s", err.Error())
	}

	if string(b) != `{"ok":1,"err_text":"","data":{"items":[{"product_id":6,"name":"Product Name","price":1234,"product_kind_id":33,"product_grp_id":113,"product_kind_name":"Kind 1","product_grp_code":"GRP1"},{"product_id":7,"name":"Product Name 2","price":1234,"product_kind_id":34,"product_grp_id":113,"product_kind_name":"Kind 2","product_grp_code":"GRP1"}]}}` {
		t.Fatalf("GET method failed to return valid JSON")
	}
}
