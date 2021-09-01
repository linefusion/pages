package typings

import (
	"reflect"

	"gopkg.in/yaml.v2"
)

type TypeField struct {
	Name string
	Data interface{}
}

type TypeParser struct {
	Fields []TypeField
	Type   reflect.Type
}

func ParseType(t reflect.Type, instance interface{}) (TypeParser, error) {
	parser := TypeParser{Type: t}
	return parser, parser.Parse(instance)
}

// Parse the given object's parserdata
func (parser *TypeParser) Parse(instance interface{}) error {
	parser.Fields = []TypeField{}
	instanceValue := reflect.ValueOf(instance)
	for fieldIndex := 0; fieldIndex < instanceValue.NumField(); fieldIndex++ {
		instanceType := instanceValue.Type().Field(fieldIndex)
		field := TypeField{
			Name: instanceType.Name,
			Data: nil,
		}

		data := reflect.New(parser.Type).Interface()
		err := yaml.Unmarshal([]byte(instanceType.Tag), data)
		if err != nil {
			return err
		}

		field.Data = data
		parser.Fields = append(parser.Fields, field)
	}
	return nil
}
