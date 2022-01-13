package definition

import "strings"

// Root contains all service definitions and will be passed to template.
type Root struct {
	// PackageName is the name of the package.
	PackageName string `json:"packageName"`
	// Services are the services described in this definition.
	Services []Service `json:"services"`
	// Objects are the structures that are used throughout this definition.
	Objects []Object `json:"objects"`
	// Imports is a map of Go imports that should be imported into Go code.
	Imports map[string]string `json:"imports"`
	// Params contains additional data parsed from command line arguments
	Params map[string]interface{} `json:"params"`
}

// Object looks up an object by name. Returns ErrNotFound error if it cannot find it.
func (d *Root) Object(name string) (*Object, error) {
	for i := range d.Objects {
		obj := &d.Objects[i]
		if obj.Name == name {
			return obj, nil
		}
	}
	return nil, ErrNotFound
}

// ObjectIsInput gets whether this object is a method input (request) type or not.
// Returns true if any method.InputObject.ObjectName matches name.
func (d *Root) ObjectIsInput(name string) bool {
	for _, service := range d.Services {
		for _, method := range service.Methods {
			if method.InputObject.ObjectName == name {
				return true
			}
		}
	}
	return false
}

// ObjectIsOutput gets whether this object is a method output (response) type or not.
// Returns true if any method.OutputObject.ObjectName matches name.
func (d *Root) ObjectIsOutput(name string) bool {
	for _, service := range d.Services {
		for _, method := range service.Methods {
			if method.OutputObject.ObjectName == name {
				return true
			}
		}
	}
	return false
}

// Service describes a service, akin to an interface in Go.
type Service struct {
	Name    string   `json:"name"`
	Methods []Method `json:"methods"`
	Comment string   `json:"comment"`
	// Metadata are typed key/value pairs extracted from the comments.
	Metadata map[string]interface{} `json:"metadata"`
}

// Method describes a method that a Service can perform.
type Method struct {
	Name           string    `json:"name"`
	NameLowerCamel string    `json:"nameLowerCamel"`
	InputObject    FieldType `json:"inputObject"`
	OutputObject   FieldType `json:"outputObject"`
	Comment        string    `json:"comment"`
	// Metadata are typed key/value pairs extracted from the comments.
	Metadata map[string]interface{} `json:"metadata"`
}

// Object describes a data structure that is part of this definition.
type Object struct {
	TypeID   string  `json:"typeID"`
	Name     string  `json:"name"`
	Imported bool    `json:"imported"`
	Fields   []Field `json:"fields"`
	Comment  string  `json:"comment"`
	// Metadata are typed key/value pairs extracted from the comments.
	Metadata map[string]interface{} `json:"metadata"`
}

// Field describes the field inside an Object.
type Field struct {
	Name           string              `json:"name"`
	NameLowerCamel string              `json:"nameLowerCamel"`
	Type           FieldType           `json:"type"`
	OmitEmpty      bool                `json:"omitEmpty"`
	Comment        string              `json:"comment"`
	Tag            string              `json:"tag"`
	ParsedTags     map[string]FieldTag `json:"parsedTags"`
	Example        interface{}         `json:"example"`
	// Metadata are typed key/value pairs extracted from the comments.
	Metadata map[string]interface{} `json:"metadata"`
}

// FieldTag is a parsed tag. For more information, see Struct Tags in Go.
type FieldTag struct {
	// Value is the value of the tag.
	Value string `json:"value"`
	// Options are the options for the tag.
	Options []string `json:"options"`
}

// FieldType holds information about the type of data that this Field stores.
type FieldType struct {
	TypeID     string `json:"typeID"`
	TypeName   string `json:"typeName"`
	ObjectName string `json:"objectName"`
	// CleanObjectName is the ObjectName with * removed for pointer types.
	CleanObjectName      string `json:"cleanObjectName"`
	ObjectNameLowerCamel string `json:"objectNameLowerCamel"`
	Multiple             bool   `json:"multiple"`
	Package              string `json:"package"`
	IsObject             bool   `json:"isObject"`
	JSType               string `json:"jsType"`
	TSType               string `json:"tsType"`
	SwiftType            string `json:"swiftType"`
}

// IsOptional returns true for pointer types (optional).
func (f FieldType) IsOptional() bool {
	return strings.HasPrefix(f.ObjectName, "*")
}
