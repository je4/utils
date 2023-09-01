package jsonutil

import (
	"encoding/json"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"reflect"
)

// Overflow is used for json fields not represented within the struct
type Overflow map[string]any

// OverflowType is needed to find the Overflow field within the struct
var OverflowType = reflect.TypeOf(Overflow{})

func UnmarshalStructWithMap(data []byte, v any) error {
	// create a map with plain json bytes per field
	var m = map[string]*JSONBytes{}
	var done = []string{}
	if err := json.Unmarshal(data, &m); err != nil {
		return errors.Wrap(err, "cannot unmarshal data to map[string]any")
	}

	valRef := reflect.ValueOf(v)

	// pointers need resove
	if valRef.Kind() == reflect.Ptr {
		valRef = reflect.Indirect(valRef)
	}
	// interfaces need resolve
	if valRef.Kind() == reflect.Interface {
		valRef = valRef.Elem()
	}
	// get the Type
	valType := valRef.Type()

	// we only deal with structs...
	if valType.Kind() != reflect.Struct {
		return errors.Errorf("value %s is not a struct", valType.String())
	}

	var overflowValue reflect.Value

	// iterate all target value fields
	for i := 0; i < valType.NumField(); i++ {
		// get the type of field i
		fldType := valType.Field(i)

		// get the value of field i
		fldValue := valRef.Field(i)

		// check for composition with Overflow
		if fldType.Name == "Overflow" && fldType.Type == OverflowType {
			overflowValue = fldValue
			continue
		}

		// ignore non exported fields
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

		keyName := jsonName
		if keyName == "" {
			keyName = fldName
		}

		// remember the json field
		done = append(done, keyName)

		// get the json bytes for the key
		valBytes, ok := m[keyName]
		// if field is not in json list ignore it
		if !ok {
			continue
		}

		// build the target value
		var target any
		switch fldType.Type.Kind() {
		case reflect.Struct:
			target = reflect.New(fldType.Type).Interface()
		case reflect.Ptr:
			target = reflect.New(fldValue.Type().Elem()).Interface()
		default:
			target = reflect.New(fldType.Type).Elem().Interface()
		}
		// unmarshal the json value into target
		// todo: figure out, how to correctly unmarshal integers
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

		// build value from target
		var newValue reflect.Value
		switch fldType.Type.Kind() {
		case reflect.Struct:
			newValue = reflect.ValueOf(target).Elem()
			/*
				case reflect.Ptr:
					newValue = reflect.ValueOf(target)
			*/
		default:
			newValue = reflect.ValueOf(target)
		}

		fldValue.Set(newValue)
	}

	if overflowValue == (reflect.Value{}) {
		return errors.Errorf("no Overflow in structure %s", valType.String())
	}

	// unmarshal all fields not represented on struct into the Overflow map
	var newMap = map[string]any{}
	for key, val := range m {
		// ignore represented values
		if slices.Contains(done, key) {
			continue
		}
		var newVal any
		if err := json.Unmarshal(*val, &newVal); err != nil {
			return errors.Wrapf(err, "cannot unmarshal field '%s'", key)
		}
		newMap[key] = newVal
	}

	// set Overflow field
	overflowValue.Set(reflect.ValueOf(newMap))

	return nil
}

func MarshalStructWithMap(v any) ([]byte, error) {
	if v == nil {
		return json.Marshal(v)
	}
	var resultMap = map[string]any{}

	value := reflect.ValueOf(v).Elem()
	valueType := value.Type()
	// iterate all struct fields
	for i := 0; i < value.NumField(); i++ {
		fldType := valueType.Field(i)
		fldValue := value.Field(i)
		// ignore unexported fields
		if !fldType.IsExported() {
			continue
		}
		fldName := fldType.Name

		// check for overlow map
		if fldName == "Overflow" && fldType.Type == OverflowType {
			/*
				if fldValue.Kind() != reflect.Map {
					return nil, errors.Errorf("field %s is of type %s - should be map", fldName, fldValue.Kind().String())
				}
			*/
			// add overflow map fields to result map
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

		// get json field tags
		jsonTag := fldType.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		jsonName, jsonOpts := parseTag(jsonTag)
		if !isValidTag(jsonName) {
			jsonName = ""
		}

		// ignore field if omitempty and zero
		if jsonOpts.Contains("omitempty") {
			if fldValue.IsZero() {
				continue
			}
		}
		newName := jsonName
		if newName == "" {
			newName = fldName
		}

		// add value to result map
		resultMap[newName] = fldValue.Interface()
	}
	data, err := json.Marshal(resultMap)
	return data, errors.WithStack(err)
}
