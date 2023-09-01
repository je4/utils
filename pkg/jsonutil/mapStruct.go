package jsonutil

import (
	"encoding/json"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"reflect"
)

type Overflow map[string]any

var OverflowType = reflect.TypeOf(Overflow{})

func UnmarshalStructWithMap(data []byte, v any) error {
	// create a map with plain json bytes per field
	var m = map[string]*JSONBytes{}
	var done = []string{}
	if err := json.Unmarshal(data, &m); err != nil {
		return errors.Wrap(err, "cannot unmarshal data to map[string]any")
	}

	valRef := reflect.ValueOf(v)

	// if its a pointer, resolve its value
	if valRef.Kind() == reflect.Ptr {
		valRef = reflect.Indirect(valRef)
	}
	// if its an interface, resolve it
	if valRef.Kind() == reflect.Interface {
		valRef = valRef.Elem()
	}
	// get the Type
	valType := valRef.Type()

	// we only deal with structs...
	if valType.Kind() != reflect.Struct {
		return errors.Errorf("value %s is not a struct", valType.String())
	}

	var mapField reflect.Value

	// iterate all target value fields
	for i := 0; i < valType.NumField(); i++ {
		// get the type of field i
		fldType := valType.Field(i)

		// get the value of field i
		fldValue := valRef.Field(i)

		// check for composition with Overflow
		if fldType.Name == "Overflow" && fldType.Type == OverflowType {
			mapField = fldValue
			continue
		}

		// if its not exported ignore it
		if !fldType.IsExported() {
			continue
		}
		// if we cannot set it, return error
		if !fldValue.CanSet() {
			return errors.Errorf("cannot set value of field '%s'", fldType.Name)
		}

		// calculate the json name of the field
		jsonTag := fldType.Tag.Get("json")
		// ignore based on json tag
		if jsonTag == "-" {
			continue
		}
		jsonName, _ := parseTag(jsonTag)
		if !isValidTag(jsonName) {
			jsonName = ""
		}

		fldName := fldType.Name
		//fldTypeString := fldType.Type.String()

		keyName := jsonName
		if keyName == "" {
			keyName = fldName
		}

		done = append(done, keyName)
		// get the json bytes for the key
		valBytes, ok := m[keyName]
		// if field is not in json list ignore it
		if !ok {
			continue
		}

		// lets build the target value
		var target any
		switch fldType.Type.Kind() {
		case reflect.Struct:
			target = reflect.New(fldType.Type).Interface()
		case reflect.Ptr:
			target = reflect.New(fldValue.Type().Elem()).Interface()
		default:
			target = reflect.New(fldType.Type).Elem().Interface()
		}
		if err := json.Unmarshal(*valBytes, &target); err != nil {
			return errors.Wrapf(err, "cannot unmarshal %s", valBytes)
		}
		// if we have a number, it will be float64 from unmarshal
		switch fldType.Type.Kind() {
		case reflect.Int:
			flVal, ok := target.(float64)
			if ok {
				target = int(flVal)
			}
		case reflect.Int8:
			flVal, ok := target.(float64)
			if ok {
				target = int8(flVal)
			}
		case reflect.Int16:
			flVal, ok := target.(float64)
			if ok {
				target = int16(flVal)
			}
		case reflect.Int32:
			flVal, ok := target.(float64)
			if ok {
				target = int32(flVal)
			}
		case reflect.Int64:
			flVal, ok := target.(float64)
			if ok {
				target = int64(flVal)
			}
		}

		// now we should have the correct value in target

		var x reflect.Value
		switch fldType.Type.Kind() {
		case reflect.Struct:
			x = reflect.ValueOf(target).Elem()
		case reflect.Ptr:
			x = reflect.ValueOf(target)
		default:
			x = reflect.ValueOf(target)
		}

		fldValue.Set(x)
	}

	if mapField == (reflect.Value{}) {
		return errors.Errorf("no Overflow in structure %s", valType.String())
	}

	var newMap = map[string]any{}
	for key, val := range m {
		if slices.Contains(done, key) {
			continue
		}
		var newVal any
		if err := json.Unmarshal(*val, &newVal); err != nil {
			return errors.Wrapf(err, "cannot unmarshal field '%s'", key)
		}
		newMap[key] = newVal
	}

	mapField.Set(reflect.ValueOf(newMap))

	return nil
}

func MarshalStructWithMap(v any) ([]byte, error) {
	if v == nil {
		return json.Marshal(v)
	}
	var resultMap = map[string]any{}

	val := reflect.ValueOf(v).Elem()
	valType := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fldType := valType.Field(i)
		fldValue := val.Field(i)
		if !fldType.IsExported() {
			continue
		}
		fldName := fldType.Name

		if fldName == "Overflow" {
			if fldValue.Kind() != reflect.Map {
				return nil, errors.Errorf("field %s is of type %s - should be map", fldName, fldValue.Kind().String())
			}
			for _, key := range fldValue.MapKeys() {
				val := fldValue.MapIndex(key)
				keyStr, ok := key.Interface().(string)
				if !ok {
					return nil, errors.Errorf("key of map %s is of type %s - should be string", fldName, key.Kind().String())
				}
				resultMap[keyStr] = val.Interface()
			}
			continue
		}

		jsonTag := fldType.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		jsonName, jsonOpts := parseTag(jsonTag)
		if !isValidTag(jsonName) {
			jsonName = ""
		}

		if jsonOpts.Contains("omitempty") {
			if fldValue.IsZero() {
				continue
			}
		}
		newName := jsonName
		if newName == "" {
			newName = fldName
		}
		resultMap[newName] = fldValue.Interface()
	}
	data, err := json.Marshal(resultMap)
	return data, errors.WithStack(err)
}
