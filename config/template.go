package config

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/creativeprojects/clog"
)

// ProfileTemplate contain the variables fed to a config template
type ProfileTemplate struct {
	Profile *Profile
}

func ResolveProfileTemplate(profile *Profile) error {
	data := ProfileTemplate{
		Profile: profile,
	}
	return resolveProfileTemplate(data, profile)
}

func resolveProfileTemplate(data ProfileTemplate, item interface{}) error {
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

	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)

		// we only consider the fields with a mapstructure tag,
		// because any other field would not be coming from the configuration file
		if key, ok := field.Tag.Lookup("mapstructure"); ok {
			if key == "" {
				continue
			}
			if valueOf.Field(i).Kind() == reflect.Ptr {
				if valueOf.Field(i).IsNil() {
					continue
				}
				// start of a new pointer to a struct
				clog.Debugf("*struct %s", key)
				err := resolveProfileTemplate(data, valueOf.Field(i).Elem().Interface())
				if err != nil {
					return err
				}
				continue
			}
			if valueOf.Field(i).Kind() == reflect.Struct {
				// start of a new struct
				clog.Debugf("struct %s", key)
				err := resolveProfileTemplate(data, valueOf.Field(i).Interface())
				if err != nil {
					return err
				}
				continue
			}
			if valueOf.Field(i).Kind() == reflect.Map {
				if valueOf.Field(i).Len() == 0 {
					continue
				}
				clog.Debugf("map %s", key)
				iter := valueOf.Field(i).MapRange()
				for iter.Next() {
					resolve, newValue, err := resolveField(data, iter.Key().String(), iter.Value())
					if err != nil {
						return err
					}
					if resolve {
						if !iter.Key().CanSet() {
							return fmt.Errorf("cannot set value of %s", iter.Key().String())
						}
						iter.Key().SetString(newValue)
					}
				}
				continue
			}
			resolve, newValue, err := resolveField(data, typeOf.Field(i).Name+">"+key, valueOf.Field(i))
			if err != nil {
				return err
			}
			if resolve {
				if !valueOf.Field(i).CanSet() {
					return fmt.Errorf("cannot set value of %s", typeOf.Field(i).Name)
				}
				valueOf.Field(i).SetString(newValue)
			}
		}
	}
	return nil
}

func resolveField(data ProfileTemplate, key string, valueOf reflect.Value) (bool, string, error) {
	if valueOf.Kind() != reflect.String {
		return false, "", nil
	}
	resolve := strings.Contains(valueOf.String(), "{{") && strings.Contains(valueOf.String(), "}}")
	clog.Debugf("(%v) %s: %v", resolve, key, valueOf.String())

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
