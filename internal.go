package crud

import (
	"fmt"
)

// initHelpers creates all the Struct2sql objects. For HTTP endpoints, it is necessary to create these first
func (c *Controller) initHelpers(newObjFunc func() interface{}, options HandlerOptions) error {
	obj := newObjFunc()

	var forceName string
	if options.ForceName != "" {
		forceName = options.ForceName
	}

	cErr := c.orm.RegisterStruct(obj, nil, false, forceName, false)
	if cErr != nil {
		return ErrController{
			Op:  "RegisterStruct",
			Err: fmt.Errorf("Error adding SQL generator: %s", cErr.Error()),
		}
	}

	if options.CreateConstructor != nil {
		cErr = c.orm.RegisterStruct(options.CreateConstructor(), obj, false, "", true)
		if cErr != nil {
			return ErrController{
				Op:  "RegisterStruct",
				Err: fmt.Errorf("Error adding SQL generator: %s", cErr.Error()),
			}
		}
	}

	if options.ReadConstructor != nil {
		cErr = c.orm.RegisterStruct(options.ReadConstructor(), obj, false, "", true)
		if cErr != nil {
			return ErrController{
				Op:  "RegisterStruct",
				Err: fmt.Errorf("Error adding SQL generator: %s", cErr.Error()),
			}
		}
	}

	if options.UpdateConstructor != nil {
		cErr = c.orm.RegisterStruct(options.UpdateConstructor(), obj, false, "", true)
		if cErr != nil {
			return ErrController{
				Op:  "RegisterStruct",
				Err: fmt.Errorf("Error adding SQL generator: %s", cErr.Error()),
			}
		}
	}

	if options.ListConstructor != nil {
		cErr = c.orm.RegisterStruct(options.ListConstructor(), obj, false, "", true)
		if cErr != nil {
			return ErrController{
				Op:  "RegisterStruct",
				Err: fmt.Errorf("Error adding SQL generator: %s", cErr.Error()),
			}
		}
	}

	return nil
}

func (c Controller) mapWithInterfacesToMapBool(m map[string]interface{}) map[string]bool {
	o := map[string]bool{}
	for k := range m {
		o[k] = true
	}
	return o
}
