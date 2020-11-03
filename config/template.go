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

// ResolveProfileTemplate loads templates from each flag and replaces the values from data
func ResolveProfileTemplate(data TemplateData, profile *Profile) error {
	return resolveProfileTemplate(data, profile)
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
			// if (valueOf.Field(fieldIndex).Kind() == reflect.Slice || valueOf.Field(fieldIndex).Kind() == reflect.Array) &&
			// 	valueOf.Field(fieldIndex).Type().Elem().Kind() == reflect.String {
			// 	for index := 0; index < valueOf.Field(fieldIndex).Len(); index++ {
			// 		// key and value are the same reflect.Value in this case
			// 		err := resolveValue(data, valueOf.Field(fieldIndex).Index(index))
			// 		if err != nil {
			// 			return fmt.Errorf("field %s[%d]: %w", typeOf.Field(fieldIndex).Name, index, err)
			// 		}
			// 	}
			// }
			// if valueOf.Field(fieldIndex).Kind() == reflect.Map {
			// 	if valueOf.Field(fieldIndex).Len() == 0 {
			// 		continue
			// 	}
			// 	iter := valueOf.Field(fieldIndex).MapRange()
			// 	for iter.Next() {
			// 		err := resolveMapValue(data, valueOf.Field(fieldIndex), iter.Key(), iter.Value())
			// 		if err != nil {
			// 			return fmt.Errorf("key \"%s\": %w", iter.Key(), err)
			// 		}
			// 	}
			// 	continue
			// }
			// err := resolveValue(data, valueOf.Field(fieldIndex))
			// if err != nil {
			// 	return fmt.Errorf("field %s: %w", typeOf.Field(fieldIndex).Name, err)
			// }
			err := resolveOtherKind(data, valueOf.Field(fieldIndex))
			if err != nil {
				return fmt.Errorf("field %s: %w", typeOf.Field(fieldIndex).Name, err)
			}
		}
	}
	return nil
}

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

func resolveMapValue(data TemplateData, sourceMap, key, value reflect.Value) error {
	currentValue := value
	if currentValue.Kind() == reflect.Interface {
		// take the actual value from inside the interface
		currentValue = currentValue.Elem()
	}

	if (currentValue.Kind() == reflect.Slice || currentValue.Kind() == reflect.Array) &&
		currentValue.Type().Elem().Kind() == reflect.Interface {
		// this is getting hairy: you cannot change any value inside the slice referenced by a map
		// so we need to create a temp reflect.Value of the slice
		temp := reflect.ValueOf(make([]interface{}, currentValue.Len()))
		reflect.Copy(temp, currentValue)
		for index := 0; index < temp.Len(); index++ {
			err := resolveValue(data, temp.Index(index))
			if err != nil {
				return fmt.Errorf("index %d: %w", index, err)
			}
		}
		// and copy this new slice into the map
		sourceMap.SetMapIndex(key, temp)
		return nil
	}

	resolve, newValue, err := resolveField(data, currentValue)
	if err != nil {
		return err
	}
	if resolve {
		sourceMap.SetMapIndex(key, reflect.ValueOf(newValue))
	}
	return nil
}

func resolveValue(data TemplateData, value reflect.Value) error {
	// our value shouldn't be of interface type - keep it in case we change our mind
	currentValue := value
	if currentValue.Kind() == reflect.Interface {
		// take the actual value from inside the interface
		currentValue = currentValue.Elem()
	}

	resolve, newValue, err := resolveField(data, currentValue)
	if err != nil {
		return err
	}
	if resolve {
		if !currentValue.CanSet() {
			if !value.CanSet() {
				// it shouldn't happen but it's still better than panic :p
				return fmt.Errorf("cannot set value of '%s', kind %v", currentValue.String(), currentValue.Kind())
			}
			// case of a value shadowed by an interface: we can't change the value itself, we must change the interface value instead
			value.Set(reflect.ValueOf(newValue))
			return nil
		}
		currentValue.SetString(newValue)
	}
	return nil
}

func resolveField(data TemplateData, valueOf reflect.Value) (bool, string, error) {
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
