package config

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"
)

// TemplateData contain the variables fed to a config template
type TemplateData struct {
	Profile    ProfileTemplateData
	Now        time.Time
	CurrentDir string
	ConfigDir  string
	Env        map[string]string
}

// ProfileTemplateData contains profile data
type ProfileTemplateData struct {
	Name string
}

func resolveProfileTemplate(data TemplateData, item interface{}) error {
	typeOf := reflect.TypeOf(item)
	valueOf := reflect.ValueOf(item)

	if typeOf.Kind() == reflect.Ptr {
		// Deference the pointer
		typeOf = typeOf.Elem()
		valueOf = valueOf.Elem()
	}

	// NumField() will panic if typeOf is not a struct
	if typeOf.Kind() != reflect.Struct {
		return fmt.Errorf("unsupported type %s, expected %s", typeOf.Kind(), reflect.Struct)
	}

	// go through all the fields of the struct
	for fieldIndex := 0; fieldIndex < typeOf.NumField(); fieldIndex++ {
		field := typeOf.Field(fieldIndex)

		// we only consider the fields with a mapstructure tag,
		// because any other field would not be coming from the configuration file
		if key, ok := field.Tag.Lookup("mapstructure"); ok {
			if key == "" {
				continue
			}
			if valueOf.Field(fieldIndex).Kind() == reflect.Ptr {
				if valueOf.Field(fieldIndex).IsNil() {
					continue
				}
				// start of a new pointer to a struct
				err := resolveProfileTemplate(data, valueOf.Field(fieldIndex).Elem().Interface())
				if err != nil {
					return err
				}
				continue
			}
			if valueOf.Field(fieldIndex).Kind() == reflect.Struct {
				// start of a new struct
				err := resolveProfileTemplate(data, valueOf.Field(fieldIndex).Interface())
				if err != nil {
					return err
				}
				continue
			}
			err := resolveOtherKind(data, valueOf.Field(fieldIndex))
			if err != nil {
				return fmt.Errorf("field %s: %w", typeOf.Field(fieldIndex).Name, err)
			}
		}
	}
	return nil
}

// resolveOtherKind resolves variable expansion for array, slice, map, interface or string
func resolveOtherKind(data TemplateData, valueOfField reflect.Value) error {
	if (valueOfField.Kind() == reflect.Slice || valueOfField.Kind() == reflect.Array) &&
		valueOfField.Type().Elem().Kind() == reflect.String {
		for index := 0; index < valueOfField.Len(); index++ {
			err := resolveValue(data, valueOfField.Index(index))
			if err != nil {
				return fmt.Errorf("index %d: %w", index, err)
			}
		}
	}
	if valueOfField.Kind() == reflect.Map {
		if valueOfField.Len() == 0 {
			return nil
		}
		iter := valueOfField.MapRange()
		for iter.Next() {
			err := resolveMapValue(data, valueOfField, iter.Key(), iter.Value())
			if err != nil {
				return fmt.Errorf("key \"%s\": %w", iter.Key(), err)
			}
		}
		return nil
	}
	err := resolveValue(data, valueOfField)
	if err != nil {
		return err
	}
	return nil
}

// resolveMapValue checks for variable expansion in a map (typically of type map[string]interface{})
// and resolves them
func resolveMapValue(data TemplateData, sourceMap, key, value reflect.Value) error {
	currentValue := value
	if currentValue.Kind() == reflect.Interface {
		// take the actual value from inside the interface
		currentValue = currentValue.Elem()
	}

	if (currentValue.Kind() == reflect.Slice || currentValue.Kind() == reflect.Array) &&
		currentValue.Type().Elem().Kind() == reflect.Interface {
		// value of type []interface{}
		for index := 0; index < currentValue.Len(); index++ {
			err := resolveValue(data, currentValue.Index(index))
			if err != nil {
				return fmt.Errorf("index %d: %w", index, err)
			}
		}
		return nil
	}

	resolve, newValue, err := resolveString(data, currentValue)
	if err != nil {
		return err
	}
	if resolve {
		sourceMap.SetMapIndex(key, reflect.ValueOf(newValue))
	}
	return nil
}

// resolveValue takes a string value, or an interface to a string value, and replaces the template content if any
func resolveValue(data TemplateData, value reflect.Value) error {
	currentValue := value
	if currentValue.Kind() == reflect.Interface {
		// take the actual value from inside the interface
		currentValue = currentValue.Elem()
	}

	resolve, newValue, err := resolveString(data, currentValue)
	if err != nil {
		return err
	}
	if resolve {
		if !currentValue.CanSet() {
			// then try the interface
			if !value.CanSet() {
				// it shouldn't happen but it's still better than panic :p
				return fmt.Errorf("cannot set value of '%s', kind %v", currentValue.String(), currentValue.Kind())
			}
			// case of a value shadowed by an interface: we can't change the value itself, we must change the interface value instead
			value.Set(reflect.ValueOf(newValue))
			return nil
		}
		// can set the value directly
		currentValue.SetString(newValue)
	}
	return nil
}

// resolveString takes the string value and check if a template needs compiled and executed.
// if the value has changed (by executing a template), it returns true and the new value
func resolveString(data TemplateData, valueOf reflect.Value) (bool, string, error) {
	if valueOf.Kind() != reflect.String {
		return false, "", nil
	}
	resolve := strings.Contains(valueOf.String(), "{{") && strings.Contains(valueOf.String(), "}}")

	tmpl, err := template.New("").Parse(valueOf.String())
	if err != nil {
		return false, "", err
	}
	buffer := &strings.Builder{}
	err = tmpl.Execute(buffer, data)
	if err != nil {
		return false, "", err
	}
	return resolve, buffer.String(), nil
}
