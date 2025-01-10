package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/go-phings/crud"
	structdbpostgres "github.com/go-phings/struct-db-postgres"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const dbDSN = "host=localhost user=cruduser password=crudpass port=54321 dbname=crud sslmode=disable"
const tblPrefix = "p_"

func main() {
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		log.Fatal("Error connecting to db")
	}

	orm := structdbpostgres.NewController(db, tblPrefix, &structdbpostgres.ControllerConfig{
		TagName: "crud",
	})
	err = orm.CreateTable(&User{})
	if err != nil {
		log.Fatalf("Error creating table: %s", err.Error())
	}

	api := crud.NewController(db, tblPrefix, &crud.ControllerConfig{
		PasswordGenerator: func(pass string) string {
			passEncrypted, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
			if err != nil {
				return ""
			}
			return string(passEncrypted)
		},
	})

	var userConstructor = func() interface{} { return &User{} }
	var userConstructorForCreate = func() interface{} { return &User_Create{} }
	var userConstructorForRead = func() interface{} { return &User_List{} }
	var userConstructorForUpdate = func() interface{} { return &User_Update{} }
	var userConstructorForList = func() interface{} { return &User_List{} }

	var userConstructorForUpdatePassword = func() interface{} { return &User_UpdatePassword{} }

	http.Handle("/users/", api.Handler("/users/", userConstructor, crud.HandlerOptions{
		CreateConstructor: userConstructorForCreate, // input fields (and JSON payload) for creating
		ReadConstructor:   userConstructorForRead,   // output fields (and JSON output) for reading
		UpdateConstructor: userConstructorForUpdate, // input fields (and JSON payload) for updating
		ListConstructor:   userConstructorForList,   // fields to appear when listing items (and JSON output)
	}))
	http.Handle("/users/password/", api.Handler("/users/password/", userConstructor, crud.HandlerOptions{
		UpdateConstructor: userConstructorForUpdatePassword, // input fields for that one updating endpoint
		Operations:        crud.OpUpdate,                    // only updating will be allowed
	}))
	log.Fatal(http.ListenAndServe(":9001", nil))
}
