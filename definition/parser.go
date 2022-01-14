package definition

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/doc"
	"go/token"
	"go/types"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/damejeras/gorpc/format"
	"github.com/fatih/structtag"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

// ErrNotFound is returned when an Object is not found.
var ErrNotFound = errors.New("not found")

// Parser parses Oto Go definition packages.
type Parser struct {
	Verbose bool

	Exclusions []string

	patterns   []string
	definition Root

	// outputObjects marks output object names.
	outputObjects map[string]struct{}
	// objects marks object names.
	objects map[string]struct{}

	// docs are the docs for extracting comments.
	docs *doc.Package
}

// NewParser makes a fresh parser using the specified patterns.
// The patterns should be the args passed into the tool (after any flags)
// and will be passed to the underlying build system.
func NewParser(patterns ...string) *Parser {
	return &Parser{
		patterns:      patterns,
		outputObjects: make(map[string]struct{}),
		objects:       make(map[string]struct{}),
	}
}

func (p *Parser) ParseWithParams(params map[string]interface{}) (*Root, error) {
	def, err := p.parse()
	if err != nil {
		return nil, err
	}

	def.Params = params

	return def, nil
}

func (p *Parser) parse() (*Root, error) {
	cfg := &packages.Config{
		Mode:  packages.NeedTypes | packages.NeedName | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedName | packages.NeedSyntax,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, p.patterns...)
	if err != nil {
		return nil, err
	}

	var excludedObjectsTypeIDs []string
	for _, pkg := range pkgs {
		p.docs, err = doc.NewFromFiles(pkg.Fset, pkg.Syntax, "")
		if err != nil {
			return nil, errors.Wrap(err, "parse docs for file")
		}

		p.definition.PackageName = pkg.Name
		scope := pkg.Types.Scope()

		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			switch item := obj.Type().Underlying().(type) {
			case *types.Interface:
				s, err := p.parseService(pkg, obj, item)
				if err != nil {
					return nil, errors.Wrap(err, "parse service")
				}

				if isInSlice(p.Exclusions, name) {
					for _, method := range s.Methods {
						excludedObjectsTypeIDs = append(excludedObjectsTypeIDs, method.InputObject.TypeID)
						excludedObjectsTypeIDs = append(excludedObjectsTypeIDs, method.OutputObject.TypeID)
					}

					continue
				}

				p.definition.Services = append(p.definition.Services, s)
			case *types.Struct:
				if err := p.parseObject(pkg, obj, item); err != nil {
					return nil, errors.Wrap(err, "parse object")
				}
			}
		}
	}

	// remove any excluded objects
	nonExcludedObjects := make([]Object, 0, len(p.definition.Objects))
	for _, object := range p.definition.Objects {
		excluded := false
		for _, excludedTypeID := range excludedObjectsTypeIDs {
			if object.TypeID == excludedTypeID {
				excluded = true

				break
			}
		}

		if excluded {
			continue
		}

		nonExcludedObjects = append(nonExcludedObjects, object)
	}

	p.definition.Objects = nonExcludedObjects
	// sort services
	sort.Slice(p.definition.Services, func(i, j int) bool {
		return p.definition.Services[i].Name < p.definition.Services[j].Name
	})
	// sort objects
	sort.Slice(p.definition.Objects, func(i, j int) bool {
		return p.definition.Objects[i].Name < p.definition.Objects[j].Name
	})

	if err := p.addOutputFields(); err != nil {
		return nil, err
	}

	return &p.definition, nil
}

func (p *Parser) parseService(pkg *packages.Package, obj types.Object, interfaceType *types.Interface) (Service, error) {
	var (
		s   Service
		err error
	)

	s.Name = obj.Name()
	s.Comment = p.commentForType(s.Name)
	s.Metadata, s.Comment, err = p.extractCommentMetadata(s.Comment)
	if err != nil {
		return s, p.wrapErr(errors.New("extract comment metadata"), pkg, obj.Pos())
	}

	if p.Verbose {
		fmt.Printf("%s ", s.Name)
	}

	l := interfaceType.NumMethods()
	for i := 0; i < l; i++ {
		m := interfaceType.Method(i)

		method, err := p.parseMethod(pkg, s.Name, m)
		if err != nil {
			return s, err
		}

		s.Methods = append(s.Methods, method)
	}

	return s, nil
}

func (p *Parser) parseMethod(pkg *packages.Package, serviceName string, methodType *types.Func) (Method, error) {
	var (
		result Method
		err    error
	)
	result.Name = methodType.Name()
	result.NameLowerCamel = format.CamelizeDown(result.Name)
	result.Comment = p.commentForMethod(serviceName, result.Name)
	result.Metadata, result.Comment, err = p.extractCommentMetadata(result.Comment)
	if err != nil {
		return result, p.wrapErr(errors.New("extract comment metadata"), pkg, methodType.Pos())
	}

	sig := methodType.Type().(*types.Signature)
	inputParams := sig.Params()
	if inputParams.Len() != 1 {
		return result, p.wrapErr(errors.New("invalid method signature: expected Method(MethodRequest) MethodResponse"), pkg, methodType.Pos())
	}

	result.InputObject, err = p.parseFieldType(pkg, inputParams.At(0))
	if err != nil {
		return result, errors.Wrap(err, "parse input object type")
	}

	outputParams := sig.Results()
	if outputParams.Len() != 1 {
		return result, p.wrapErr(errors.New("invalid method signature: expected Method(MethodRequest) MethodResponse"), pkg, methodType.Pos())
	}

	result.OutputObject, err = p.parseFieldType(pkg, outputParams.At(0))
	if err != nil {
		return result, errors.Wrap(err, "parse output object type")
	}

	p.outputObjects[result.OutputObject.TypeName] = struct{}{}

	return result, nil
}

// parseObject parses a struct type and adds it to the Root.
func (p *Parser) parseObject(pkg *packages.Package, o types.Object, v *types.Struct) error {
	var (
		obj Object
		err error
	)
	obj.Name = o.Name()
	obj.Comment = p.commentForType(obj.Name)
	obj.Metadata, obj.Comment, err = p.extractCommentMetadata(obj.Comment)
	if err != nil {
		return p.wrapErr(errors.New("extract comment metadata"), pkg, o.Pos())
	}
	if _, found := p.objects[obj.Name]; found {
		// if this has already been parsed, skip it
		return nil
	}
	if o.Pkg().Name() != pkg.Name {
		obj.Imported = true
	}
	typ := v.Underlying()
	st, ok := typ.(*types.Struct)
	if !ok {
		return p.wrapErr(errors.New(obj.Name+" must be a struct"), pkg, o.Pos())
	}
	obj.TypeID = o.Pkg().Path() + "." + obj.Name
	obj.Fields = []Field{}
	for i := 0; i < st.NumFields(); i++ {
		field, err := p.parseField(pkg, obj.Name, st.Field(i), st.Tag(i))
		if err != nil {
			return err
		}
		field.Tag = v.Tag(i)
		field.ParsedTags, err = p.parseTags(field.Tag)
		if err != nil {
			return errors.Wrap(err, "parse field tag")
		}
		obj.Fields = append(obj.Fields, field)
	}
	p.definition.Objects = append(p.definition.Objects, obj)
	p.objects[obj.Name] = struct{}{}
	return nil
}

func (p *Parser) parseTags(tag string) (map[string]FieldTag, error) {
	tags, err := structtag.Parse(tag)
	if err != nil {
		return nil, err
	}
	fieldTags := make(map[string]FieldTag)
	for _, tag := range tags.Tags() {
		fieldTags[tag.Key] = FieldTag{
			Value:   tag.Name,
			Options: tag.Options,
		}
	}
	return fieldTags, nil
}

func (p *Parser) parseField(pkg *packages.Package, objectName string, v *types.Var, tag string) (Field, error) {
	var f Field
	f.Name = v.Name()
	f.NameLowerCamel = format.CamelizeDown(f.Name)
	// if it has a json tag, use that as the NameJSON.
	if tag != "" {
		fieldTag := reflect.StructTag(tag)
		jsonTag := fieldTag.Get("json")
		if jsonTag != "" {
			f.NameLowerCamel = strings.Split(jsonTag, ",")[0]
		}
	}
	f.Comment = p.commentForField(objectName, f.Name)
	f.Metadata = map[string]interface{}{}
	if !v.Exported() {
		return f, p.wrapErr(errors.New(f.Name+" must be exported"), pkg, v.Pos())
	}
	var err error
	f.Metadata, f.Comment, err = p.extractCommentMetadata(f.Comment)
	if err != nil {
		return f, p.wrapErr(errors.New("extract comment metadata"), pkg, v.Pos())
	}
	if example, ok := f.Metadata["example"]; ok {
		f.Example = example
	}
	f.Type, err = p.parseFieldType(pkg, v)
	if err != nil {
		return f, errors.Wrap(err, "parse type")
	}
	return f, nil
}

func (p *Parser) parseFieldType(pkg *packages.Package, obj types.Object) (FieldType, error) {
	var ftype FieldType
	pkgPath := pkg.PkgPath
	resolver := func(other *types.Package) string {
		if other.Name() != pkg.Name {
			if p.definition.Imports == nil {
				p.definition.Imports = make(map[string]string)
			}
			p.definition.Imports[other.Path()] = other.Name()
			ftype.Package = other.Path()
			pkgPath = other.Path()
			return other.Name()
		}
		return "" // no package prefix
	}

	typ := obj.Type()
	if slice, ok := obj.Type().(*types.Slice); ok {
		typ = slice.Elem()
		ftype.Multiple = true
	}

	originalTyp := typ
	pointerType, isPointer := typ.(*types.Pointer)
	if isPointer {
		typ = pointerType.Elem()
		isPointer = true
	}

	if named, ok := typ.(*types.Named); ok {
		if structure, ok := named.Underlying().(*types.Struct); ok {
			if err := p.parseObject(pkg, named.Obj(), structure); err != nil {
				return ftype, err
			}
			ftype.IsObject = true
		}
	}

	// disallow nested structs
	switch typ.(type) {
	case *types.Struct:
		return ftype, p.wrapErr(errors.New("nested structs not supported (create another type instead)"), pkg, obj.Pos())
	}
	ftype.TypeName = types.TypeString(originalTyp, resolver)
	ftype.ObjectName = types.TypeString(originalTyp, func(other *types.Package) string { return "" })
	ftype.ObjectNameLowerCamel = format.CamelizeDown(ftype.ObjectName)
	ftype.TypeID = pkgPath + "." + ftype.ObjectName
	ftype.CleanObjectName = strings.TrimPrefix(ftype.TypeName, "*")
	ftype.IsPointer = isPointer
	ftype.TSType = ftype.CleanObjectName
	ftype.JSType = ftype.CleanObjectName
	ftype.SwiftType = ftype.CleanObjectName
	if ftype.IsObject {
		ftype.JSType = "object"
		//ftype.SwiftType = "Any"
	} else {
		switch ftype.CleanObjectName {
		case "interface{}":
			ftype.JSType = "any"
			ftype.SwiftType = "Any"
			ftype.TSType = "object"
			ftype.PHPType = "mixed"
		case "map[string]interface{}":
			ftype.JSType = "object"
			ftype.TSType = "object"
			ftype.SwiftType = "Any"
			ftype.PHPType = "array"
		case "string":
			ftype.JSType = "string"
			ftype.SwiftType = "String"
			ftype.TSType = "string"
			ftype.PHPType = "string"
		case "bool":
			ftype.JSType = "boolean"
			ftype.SwiftType = "Bool"
			ftype.TSType = "boolean"
			ftype.PHPType = "bool"
		case "int", "int16", "int32", "int64",
			"uint", "uint16", "uint32", "uint64",
			"float32", "float64":
			ftype.JSType = "number"
			ftype.SwiftType = "Double"
			ftype.TSType = "number"
			ftype.PHPType = "float"
		}
	}

	return ftype, nil
}

// addOutputFields adds built-in fields to the response objects
// mentioned in p.outputObjects.
func (p *Parser) addOutputFields() error {
	errorField := Field{
		OmitEmpty:      true,
		Name:           "Error",
		NameLowerCamel: "error",
		Comment:        "Error is string explaining what went wrong. Empty if everything was fine.",
		Type: FieldType{
			TypeName:  "string",
			JSType:    "string",
			SwiftType: "String",
			TSType:    "string",
		},
		Metadata: map[string]interface{}{},
		Example:  "something went wrong",
	}
	for typeName := range p.outputObjects {
		obj, err := p.definition.Object(typeName)
		if err != nil {
			// skip if we can't find it - it must be excluded
			continue
		}
		obj.Fields = append(obj.Fields, errorField)
	}
	return nil
}

func (p *Parser) wrapErr(err error, pkg *packages.Package, pos token.Pos) error {
	position := pkg.Fset.Position(pos)
	return errors.Wrap(err, position.String())
}

func isInSlice(slice []string, s string) bool {
	for i := range slice {
		if slice[i] == s {
			return true
		}
	}
	return false
}

func (p *Parser) lookupType(name string) *doc.Type {
	for i := range p.docs.Types {
		if p.docs.Types[i].Name == name {
			return p.docs.Types[i]
		}
	}
	return nil
}

func (p *Parser) commentForType(name string) string {
	typ := p.lookupType(name)
	if typ == nil {
		return ""
	}
	return cleanComment(typ.Doc)
}

func (p *Parser) commentForMethod(service, method string) string {
	typ := p.lookupType(service)
	if typ == nil {
		return ""
	}
	spec, ok := typ.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return ""
	}
	iface, ok := spec.Type.(*ast.InterfaceType)
	if !ok {
		return ""
	}
	var m *ast.Field
outer:
	for i := range iface.Methods.List {
		for _, name := range iface.Methods.List[i].Names {
			if name.Name == method {
				m = iface.Methods.List[i]
				break outer
			}
		}
	}
	if m == nil {
		return ""
	}
	return cleanComment(m.Doc.Text())
}

func (p *Parser) commentForField(typeName, field string) string {
	typ := p.lookupType(typeName)
	if typ == nil {
		return ""
	}
	spec, ok := typ.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return ""
	}
	obj, ok := spec.Type.(*ast.StructType)
	if !ok {
		return ""
	}
	var f *ast.Field
outer:
	for i := range obj.Fields.List {
		for _, name := range obj.Fields.List[i].Names {
			if name.Name == field {
				f = obj.Fields.List[i]
				break outer
			}
		}
	}
	if f == nil {
		return ""
	}
	return cleanComment(f.Doc.Text())
}

func cleanComment(s string) string {
	return strings.TrimSpace(s)
}

// metadataCommentRegex is the regex to pull key value metadata
// used since we can't simply trust lines that contain a colon
var metadataCommentRegex = regexp.MustCompile(`^.*: .*`)

// extractCommentMetadata extracts key value pairs from the comment.
// It returns a map of metadata, and the
// remaining comment string.
// Metadata fields should succeed the comment string.
func (p *Parser) extractCommentMetadata(comment string) (map[string]interface{}, string, error) {
	var lines []string
	var metadata = make(map[string]interface{})
	s := bufio.NewScanner(strings.NewReader(comment))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if metadataCommentRegex.MatchString(line) {
			line = strings.TrimSpace(line)
			if line == "" {
				return metadata, strings.Join(lines, "\n"), nil
			}
			// SplitN is being used to ensure that colons can exist
			// in values by only splitting on the first colon in the line
			splitLine := strings.SplitN(line, ": ", 2)
			key := splitLine[0]
			value := strings.TrimSpace(splitLine[1])
			var val interface{}
			if err := json.Unmarshal([]byte(value), &val); err != nil {
				if p.Verbose {
					fmt.Printf("(skipping) failed to marshal JSON value (%s): %s\n", err, value)
				}
				continue
			}
			metadata[key] = val
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return metadata, strings.Join(lines, "\n"), nil
}

// ParseParams returns a map of data parsed from the params string.
func ParseParams(s string) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if s == "" {
		// empty map for an empty string
		return params, nil
	}
	pairs := strings.Split(s, ",")
	for i := range pairs {
		pair := strings.TrimSpace(pairs[i])
		segs := strings.Split(pair, ":")
		if len(segs) != 2 {
			return nil, errors.New("malformed params")
		}
		params[strings.TrimSpace(segs[0])] = strings.TrimSpace(segs[1])
	}
	return params, nil
}
