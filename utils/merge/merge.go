// Package merge provides a utility for merging structs of different types using reflection
// Borrowed from https://stackoverflow.com/a/45258790/1562480 written by Pavlo Strokov
package merge

import ("reflect")

type Translator func(value interface{}) (interface{}, error)

type Bind struct {
	From, To   string
	Translator Translator
}

func Merge(from, to interface{}, bindings ...Bind) error {
	binds := make(map[string]Bind)
	for _, bind := range bindings {
		binds[bind.From] = bind
	}
	fromValue := reflect.ValueOf(from)
	if fromValue.Kind() == reflect.Ptr {
		fromValue = fromValue.Elem()
	}
	fromFields := fieldValuesByNames(fromValue)
	toFields := fieldValuesByNames(reflect.ValueOf(to).Elem())
	for fromFieldName, fromFieldValue := range fromFields {
		if fieldBind, found := binds[fromFieldName]; found {
			if toFieldValue, found := toFields[fieldBind.To]; found {
				translated, err := fieldBind.Translator(fromFieldValue.Interface())
				if err != nil {
					return err
				}
				toFieldValue.Set(reflect.ValueOf(translated))
			}
		} else {
			if toFieldValue, found := toFields[fromFieldName]; found && toFieldValue.CanSet() {
				toFieldValue.Set(fromFieldValue)
			}
		}
	}
	return nil
}

func fieldValuesByNames(value reflect.Value) map[string]reflect.Value {
	res := make(map[string]reflect.Value)
	for fieldIndex := 0; fieldIndex < value.NumField(); fieldIndex++ {
		fieldValue := value.Field(fieldIndex)
		fieldName := value.Type().Field(fieldIndex).Name
		res[fieldName] = fieldValue
	}
	return res
}