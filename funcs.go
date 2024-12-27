package crud

import "reflect"

func getStructName(u interface{}) string {
	v := reflect.ValueOf(u)
	i := reflect.Indirect(v)
	s := i.Type()
	return s.Name()
}
