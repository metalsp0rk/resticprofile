package config

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/creativeprojects/clog"
)

// TemplateData contain the variables fed to a config template
type TemplateData struct {
	Profile    ProfileTemplateData
	Now        time.Time
	CurrentDir string
	ConfigDir  string
}

// ProfileTemplateData contains profile data
type ProfileTemplateData struct {
	Name string
}

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
			if (valueOf.Field(fieldIndex).Kind() == reflect.Slice || valueOf.Field(fieldIndex).Kind() == reflect.Array) &&
				valueOf.Field(fieldIndex).Type().Elem().Kind() == reflect.String {
				for index := 0; index < valueOf.Field(fieldIndex).Len(); index++ {
					// key and value are the same reflect.Value in this case
					err := resolveValue(data, valueOf.Field(fieldIndex).Index(index))
					if err != nil {
						return fmt.Errorf("field %s[%d]: %w", typeOf.Field(fieldIndex).Name, index, err)
					}
				}
			}
			if valueOf.Field(fieldIndex).Kind() == reflect.Map {
				if valueOf.Field(fieldIndex).Len() == 0 {
					continue
				}
				iter := valueOf.Field(fieldIndex).MapRange()
				for iter.Next() {
					err := resolveMapValue(data, valueOf.Field(fieldIndex), iter.Key(), iter.Value())
					if err != nil {
						return fmt.Errorf("key \"%s\": %w", iter.Key(), err)
					}
				}
				continue
			}
			err := resolveValue(data, valueOf.Field(fieldIndex))
			if err != nil {
				return fmt.Errorf("field %s: %w", typeOf.Field(fieldIndex).Name, err)
			}
		}
	}
	return nil
}

func resolveMapValue(data TemplateData, sourceMap, key, value reflect.Value) error {
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
		sourceMap.SetMapIndex(key, reflect.ValueOf(newValue))
	}
	return nil
}

func resolveValue(data TemplateData, value reflect.Value) error {
	// our value shouldn't be of interface type - keep it in case we change our mind
	currentValue := value
	if currentValue.Kind() == reflect.Interface {
		clog.Debugf("found interface value: %v", currentValue)
		// take the actual value from inside the interface
		currentValue = currentValue.Elem()
	}

	resolve, newValue, err := resolveField(data, currentValue)
	if err != nil {
		return err
	}
	if resolve {
		if !currentValue.CanSet() {
			// it shouldn't happen but it's still better than panic :p
			return fmt.Errorf("cannot set value of '%s', kind %v", currentValue.String(), currentValue.Kind())
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
