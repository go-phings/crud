package crud

import (
	"database/sql"
)

// Controller is the main component that gets and saves objects in the database and generates CRUD HTTP handler
// that can be attached to an HTTP server.
type Controller struct {
	orm      ORM
	tagName  string
	passFunc func(string) string
}

type ControllerConfig struct {
	TagName           string
	PasswordGenerator func(string) string
	ORM               ORM
}

type ContextValue string

// NewController returns new Controller object
func NewController(dbConn *sql.DB, tblPrefix string, cfg *ControllerConfig) *Controller {
	c := &Controller{}

	c.tagName = "crud"
	if cfg != nil && cfg.TagName != "" {
		c.tagName = cfg.TagName
	}

	if cfg != nil && cfg.PasswordGenerator != nil {
		c.passFunc = cfg.PasswordGenerator
	}

	if cfg != nil && cfg.ORM != nil {
		c.orm = cfg.ORM
	} else {
		c.orm = newWrappedStruct2db(dbConn, tblPrefix, c.tagName)
	}

	return c
}
