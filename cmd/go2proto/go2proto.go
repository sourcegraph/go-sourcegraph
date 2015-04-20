package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/serenize/snaker"
)

var (
	writeFiles = flag.Bool("w", false, "write result to .proto files instead of stdout")
	typeFilter = flag.String("t", "", "type filter: regexp specifying which Go types to convert")
	outFile    = flag.String("o", "", "output .proto file (default: PKG.proto where PKG is pkg name from -i)")

	fset = token.NewFileSet()
)

const (
	docWrap = 80
	indent  = "\t"
	comment = "// "
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	var (
		dir   string
		files []string
	)
	switch flag.NArg() {
	case 0:
		dir = "."
	case 1:
		path := flag.Arg(0)
		if fi, err := os.Stat(path); err != nil {
			log.Fatal(err)
		} else if fi.IsDir() {
			dir = path
		} else {
			files = []string{filepath.Base(path)}
			dir = filepath.Dir(path)
		}
	default:
		// ensure all files listed are in same dir
		for _, f := range flag.Args() {
			if fi, err := os.Stat(f); err != nil {
				log.Fatal(err)
			} else if fi.IsDir() {
				log.Fatalf("Error: when specifying multiple args, all of them must be files. (%s is a directory.)", f)
			}
			if dir != "" && filepath.Dir(f) != dir {
				log.Fatalf("Error: all files specified must be in the same directory (%s != %s).", dir, filepath.Dir(f))
			}
			dir = filepath.Dir(f)
			files = append(files, filepath.Base(f))
		}
	}

	bpkg, err := build.ImportDir(dir, 0)
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		// default to using non-test, non-ignored .go files
		files = bpkg.GoFiles
	}

	if *outFile == "" {
		tmp := bpkg.Name + ".proto"
		outFile = &tmp
	}

	goFilesNoTest := func(fi os.FileInfo) bool {
		name := fi.Name()
		for _, f := range files {
			if f == name {
				return true
			}
		}
		return false
	}
	pkgs, err := parser.ParseDir(fset, dir, goFilesNoTest, parser.AllErrors|parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("Error: expected exactly 1 Go package in %s, found %d.", dir, len(pkgs))
	}
	var pkg *ast.Package
	for _, pkg2 := range pkgs {
		pkg = pkg2
	}

	doc := doc.New(pkg, bpkg.ImportPath, 0)
	if err != nil {
		log.Fatal(err)
	}

	typeFilterRE, err := regexp.Compile(*typeFilter)
	if err != nil {
		log.Fatal(err)
	}

	b := protoBuilder{
		pkg:        pkg,
		doc:        doc,
		dir:        dir,
		typeFilter: typeFilterRE.MatchString,
	}
	b.build()
	b.analyze()
	if err := b.write(*writeFiles); err != nil {
		log.Fatalf("Error writing protobuf for package in %s: %s.", dir, err)
	}
}

const (
	gogoExtsProto      = "gogoproto/gogo.proto"
	timestampProtoFile = "timestamp.proto"
)

type protoFile struct {
	pkg      string
	imports  []string
	options  []string
	messages []*protoMessage
	services []*protoService
}

func (f *protoFile) analyze() error {
	for _, msg := range f.messages {
		msg.file = f
		if err := msg.analyze(); err != nil {
			return fmt.Errorf("message %s: %s", msg.name, err)
		}
	}
	return nil
}

func (f *protoFile) addImport(imp string) {
	// don't add duplicate imports
	for _, imp2 := range f.imports {
		if imp2 == imp {
			return
		}
	}
	f.imports = append(f.imports, imp)
	sort.Strings(f.imports)
}

func (f *protoFile) write(w io.Writer) error {
	// if f.options == nil {
	// 	f.addImport("gogoproto/gogo.proto")
	// 	f.options = []string{
	// 		"(gogoproto.populate_all) = true;",
	// 		"(gogoproto.testgen_all) = true;",
	// 		"(gogoproto.benchgen_all) = true;",
	// 		"(gogoproto.equal_all) = true;",
	// 	}
	// }

	fmt.Fprintln(w, `syntax = "proto3";`)
	fmt.Fprintf(w, "package %s;\n", f.pkg)
	fmt.Fprintln(w)
	for _, imp := range f.imports {
		fmt.Fprintf(w, "import %q;\n", imp)
	}
	if len(f.imports) != 0 {
		fmt.Fprintln(w)
	}
	for _, opt := range f.options {
		fmt.Fprintf(w, "option %s;\n", opt)
	}
	if len(f.options) != 0 {
		fmt.Fprintln(w)
	}
	for _, msg := range f.messages {
		fmt.Fprintln(w)
		if err := msg.write(w); err != nil {
			return fmt.Errorf("message %s: %s", msg.name, err)
		}
	}
	if len(f.messages) != 0 {
		fmt.Fprintln(w)
	}
	for _, svc := range f.services {
		fmt.Fprintln(w)
		if err := svc.write(w); err != nil {
			return fmt.Errorf("service %s: %s", svc.name, err)
		}
	}
	return nil
}

type protoMessage struct {
	name   string
	doc    string
	fields []*protoField

	file *protoFile // containing file
}

func (m *protoMessage) equal(other *protoMessage) bool {
	if (m == nil) != (other == nil) {
		return false
	}
	var a, b bytes.Buffer
	if err := m.write(&a); err != nil {
		panic(err)
	}
	if err := m.write(&b); err != nil {
		panic(err)
	}
	return a.String() == b.String()
}

func (m *protoMessage) analyze() error {
	for _, field := range m.fields {
		field.file = m.file
		if err := field.analyze(); err != nil {
			return fmt.Errorf("field %s: %s", field.name, err)
		}
	}
	return nil
}

func (m *protoMessage) write(w io.Writer) error {
	docToText(w, m.doc, comment)
	fmt.Fprintf(w, "message %s {\n", m.name)
	for i, field := range m.fields {
		if i != 0 && (field.doc != "" || m.fields[i-1].doc != "") {
			fmt.Fprintln(w)
		}
		if err := field.write(w); err != nil {
			return fmt.Errorf("field %s: %s", field.name, err)
		}
	}
	fmt.Fprintln(w, "}")
	return nil
}

type protoField struct {
	name string
	doc  string
	tag  int

	protoFieldType

	// gogoproto extensions
	customName string
	embedded   bool
	moreTags   []string // joined by spaces

	file *protoFile // containing file
}

type protoFieldType struct {
	typeName string
	repeated bool
	optional bool

	// gogoproto extensions
	customType  string
	nonNullable bool

	origin string // .proto file where this type is defined (externally defined types only)
}

func (f *protoField) analyze() error {
	_, imports := f.extensions()
	for _, imp := range imports {
		f.file.addImport(imp)
	}
	if f.origin != "" {
		f.file.addImport(f.origin)
	}
	return nil
}

func (f *protoField) extensions() (exts []string, imports []string) {
	formatExt := func(name string, val interface{}) string {
		return fmt.Sprintf("(%s) = %#v", name, val)
	}

	var needsGogoExtsImport bool
	if f.customName != "" {
		exts = append(exts, formatExt("gogoproto.customname", f.customName))
		needsGogoExtsImport = true
	}
	if f.customType != "" {
		exts = append(exts, formatExt("gogoproto.customtype", f.customType))
		needsGogoExtsImport = true
	}
	if f.nonNullable {
		exts = append(exts, formatExt("gogoproto.nullable", false))
		needsGogoExtsImport = true
	}
	if len(f.moreTags) > 0 {
		exts = append(exts, formatExt("gogoproto.moretags", strings.Join(f.moreTags, " ")))
		needsGogoExtsImport = true
	}
	if f.embedded {
		exts = append(exts, formatExt("gogoproto.embed", f.embedded))
		needsGogoExtsImport = true
	}

	if needsGogoExtsImport {
		imports = append(imports, gogoExtsProto)
	}

	return
}

func (f *protoField) write(w io.Writer) error {
	// validations
	if f.optional && f.repeated {
		return errors.New("field may not be both optional and repeated")
	}

	docToText(w, f.doc, indent+comment)
	fmt.Fprint(w, indent)
	if f.optional {
		fmt.Fprint(w, "optional ")
	}
	if f.repeated {
		fmt.Fprint(w, "repeated ")
	}
	fmt.Fprint(w, f.typeName, " ", f.name, " = ", f.tag)

	if exts, _ := f.extensions(); len(exts) != 0 {
		fmt.Fprint(w, " [", strings.Join(exts, ", "), "]")
	}

	fmt.Fprintln(w, ";")

	return nil
}

type protoService struct {
	name string
	doc  string

	methods []*protoMethod

	file *protoFile
}

func (s *protoService) write(w io.Writer) error {
	docToText(w, s.doc, comment)
	fmt.Fprintf(w, "service %s {\n", s.name)
	for i, m := range s.methods {
		if i != 0 && (m.doc != "" || s.methods[i-1].doc != "") {
			fmt.Fprintln(w)
		}
		if err := m.write(w); err != nil {
			return fmt.Errorf("method %s: %s", m.name, err)
		}
	}
	fmt.Fprintln(w, "}")
	return nil
}

type protoMethod struct {
	name    string
	doc     string
	arg     string
	returns string

	file *protoFile
}

func (m *protoMethod) write(w io.Writer) error {
	docToText(w, m.doc, indent+comment)
	fmt.Fprint(w, indent)
	fmt.Fprintf(w, "rpc %s(%s) returns (%s);\n", m.name, m.arg, m.returns)
	return nil
}

type protoBuilder struct {
	pkg *ast.Package
	doc *doc.Package
	dir string

	typeFilter func(string) bool

	protoFiles map[string]*protoFile
}

func (b *protoBuilder) file(goFile string) *protoFile {
	if b.protoFiles == nil {
		b.protoFiles = map[string]*protoFile{}
	}

	// uncomment to output to multiple .proto files:
	//
	// name := goFile[:len(goFile)-len(".go")] + ".proto"

	name := *outFile

	if _, ok := b.protoFiles[name]; !ok {
		b.protoFiles[name] = &protoFile{pkg: b.doc.Name}
	}
	return b.protoFiles[name]
}

func (b *protoBuilder) build() {
	skipStructType := func(x *ast.TypeSpec) bool {
		name := x.Name.Name
		file := filepath.Base(fset.Position(x.Pos()).Filename)
		return !ast.IsExported(name) || strings.HasPrefix(name, "Err") || strings.HasSuffix(name, "Error") ||
			(file == "client.go" && (name == "Client" || name == "HTTPResponse")) ||
			strings.HasPrefix(name, "Mock")
	}

	for _, typ := range b.doc.Types {
		if !b.typeFilter(typ.Name) {
			continue
		}
		tspec := typ.Decl.Specs[0].(*ast.TypeSpec)
		if skipStructType(tspec) {
			continue
		}

		file := b.file(filepath.Base(fset.Position(tspec.Pos()).Filename))

		// structs become messages
		if t, ok := tspec.Type.(*ast.StructType); ok {
			b.buildMessage(file, typ.Name, typ.Doc, t)
		}

		// TODO(sqs): get Consts and convert to enums

		// interfaces become services
		if t, ok := tspec.Type.(*ast.InterfaceType); ok {
			if stripService := true; stripService {
				typ.Name = strings.TrimSuffix(typ.Name, "Service")
			}
			b.buildService(file, typ.Name, typ.Doc, t)
		}
	}
}

func (b *protoBuilder) buildMessage(file *protoFile, name, doc string, t *ast.StructType) {
	msg := &protoMessage{
		name: name,
		doc:  doc,
		file: file,
	}

	fieldTag := 1
	for _, goField := range t.Fields.List {
		// treat embedded types as named fields
		var embedded bool
		if len(goField.Names) == 0 {
			goField.Names = []*ast.Ident{ast.NewIdent(embeddedTypeName(goField.Type))}
			embedded = true
		}

		for _, name := range goField.Names {
			field := &protoField{
				name:     camelToSnake(name.Name),
				doc:      goField.Doc.Text(), // TODO(sqs): doc is duplicated when len(f.Names) > 1
				tag:      fieldTag,
				embedded: embedded,
				file:     file,
			}
			fieldTag++

			if needsExplicitName(name.Name) {
				field.customName = name.Name
			}

			field.protoFieldType = equivProtoType(goField.Type)

			if goField.Tag != nil {
				stag := reflect.StructTag(strings.Trim(goField.Tag.Value, "`"))
				if v := stag.Get("db"); v != "" && v != camelToSnake(name.Name) {
					field.moreTags = append(field.moreTags, fmt.Sprintf("db:%q", v))
				}
				for _, key := range []string{"url", "schema"} {
					if v := stag.Get(key); v != "" {
						field.moreTags = append(field.moreTags, fmt.Sprintf("%s:%q", key, v))
					}
				}
			}

			msg.fields = append(msg.fields, field)
		}
	}

	// Add the new message to the file (check if a different one with
	// the same name exists first).
	for _, msg2 := range file.messages {
		if msg.name == msg2.name && !msg.equal(msg2) {
			log.Fatalf("Error: 2 messages named %q with conflicting definitions.", msg.name)
		}
	}
	file.messages = append(file.messages, msg)
}

func (b *protoBuilder) buildService(file *protoFile, name, doc string, t *ast.InterfaceType) {
	svc := &protoService{
		name: name,
		doc:  doc,
		file: file,
	}
	file.services = append(file.services, svc)

	for _, goMethod := range t.Methods.List {
		if len(goMethod.Names) == 0 {
			p := fset.Position(goMethod.Pos())
			log.Printf("# warning: (%s).%s @ %s:%d: interface embedding is not supported for protobuf service generation", name, astString(goMethod.Type), p.Filename, p.Line)
		}

		for _, name := range goMethod.Names {
			b.buildServiceMethod(svc, name.Name, goMethod.Doc.Text(), goMethod.Type.(*ast.FuncType))
		}
	}
}

func (b *protoBuilder) buildServiceMethod(svc *protoService, name string, doc string, typ *ast.FuncType) {
	m := &protoMethod{
		name: name,
		doc:  doc,
		file: svc.file,
	}
	svc.methods = append(svc.methods, m)

	// Create new message types for the arg/return values unless they
	// consist of exactly 1 existing message type. The suffix is used
	// if a new message type is created unless it is a list or
	// something similar (e.g., []*Foo will become FooList, not
	// BarResult).
	protoSingleMessageType := func(fl *ast.FieldList, suffix string) string {
		if len(fl.List) == 1 {
			t := equivProtoType(fl.List[0].Type)
			if t.repeated {
				// create new "XxxList" message type
				elt := fl.List[0].Type
				goType := &ast.StructType{
					Fields: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent(nounForType(elt))},
								Type:  &ast.ArrayType{Elt: elt},
							},
						},
					},
				}
				synthedName := t.typeName + "List"
				b.buildMessage(svc.file, synthedName, "", goType)
				return synthedName
			}
			// use existing
			return t.typeName
		}
		// Create a new wrapper type.
		for _, f := range fl.List {
			if len(f.Names) == 0 {
				// Ensure all the fields have names so they aren't
				// treated as embedded types (the names are discarded
				// later).
				f.Names = []*ast.Ident{ast.NewIdent(nounForType(f.Type))}
			}
			for _, name := range f.Names {
				// Uppercase all of the names because they are going
				// to become fields in a synthesized struct (so they
				// should be "exported").
				name.Name = strings.ToUpper(name.Name[0:1]) + name.Name[1:]
			}
		}
		goType := &ast.StructType{Fields: fl}
		synthedName := stripServiceRelatedSuffix(svc.name) + name + suffix
		b.buildMessage(svc.file, synthedName, "", goType)
		return synthedName
	}

	// Remove a leading "context.Context" param because that is added
	// automatically by grpc.
	if args := typ.Params.List; len(args) > 0 && astString(args[0].Type) == "context.Context" {
		typ.Params.List = args[1:]
	}
	m.arg = protoSingleMessageType(typ.Params, "Op")

	// Remove a trailing "error" return because errors are handled
	// out-of-band in protobuf RPC.
	if rs := typ.Results.List; len(rs) > 0 && astString(rs[len(rs)-1].Type) == "error" {
		typ.Results.List = rs[:len(rs)-1]
	}
	m.returns = protoSingleMessageType(typ.Results, "Result")
}

func (b *protoBuilder) analyze() error {
	for name, f := range b.protoFiles {
		if err := f.analyze(); err != nil {
			return fmt.Errorf("analyzing %s: %s", name, err)
		}
	}
	return nil
}

func (b *protoBuilder) write(writeFiles bool) error {
	filenames := make([]string, 0, len(b.protoFiles))
	for name := range b.protoFiles {
		filenames = append(filenames, name)
	}
	sort.Strings(filenames)
	for _, name := range filenames {
		log.Printf("# %s", name)

		var w io.Writer
		if writeFiles {
			f, err := os.Create(name)
			if err != nil {
				return err
			}
			defer f.Close()
		} else {
			w = os.Stdout
		}

		if err := b.protoFiles[name].write(w); err != nil {
			return fmt.Errorf("writing %s: %s", name, err)
		}
	}
	return nil
}

// snakeNameSingleTerms lets us avoid converting "VCS" to "v_c_s",
// etc., in camelToSnake.
var snakeNameSingleTerms = map[string]struct{}{
	"VCS":    struct{}{},
	"GitHub": struct{}{},
}

func camelToSnake(name string) string {
	for term := range snakeNameSingleTerms {
		name = strings.Replace(name, term, term[0:1]+strings.ToLower(term[1:]), -1)
	}
	return snaker.CamelToSnake(name)
}

// needsExplicitName reports whether an explicit field name must be
// specified (using the gogoproto.customname extension) to ensure the
// generated Go field name will match the original Go input field
// name. generator.CamelCase is the func used by protoc-gen-go (and
// gogoproto).
func needsExplicitName(name string) bool {
	return generator.CamelCase(camelToSnake(name)) != name
}

const timestampTypeName = "Timestamp"

// customTypeMapping is consulted by equivProtoType when it can't
// automatically determine the correct protobuf type to use for a Go
// type expr.
var customTypeMapping = map[string]protoFieldType{
	"time.Time":          protoFieldType{typeName: timestampTypeName, nonNullable: true, origin: timestampProtoFile},
	"db_common.NullTime": protoFieldType{typeName: timestampTypeName, optional: true, origin: timestampProtoFile},
	"template.HTML":      protoFieldType{typeName: "string", customType: "template.HTML"},
	"graph.Def":          protoFieldType{typeName: "graph.Def", origin: "../../srclib/graph/def.proto"},
	"graph.DefKey":       protoFieldType{typeName: "graph.DefKey", origin: "../../srclib/graph/def.proto"},
	"graph.Ref":          protoFieldType{typeName: "graph.Ref", origin: "../../srclib/graph/ref.proto"},
	"vcs.Commit":         protoFieldType{typeName: "vcs.Commit", origin: "vcs.proto"},
}

func equivProtoType(t ast.Expr) protoFieldType {
	switch t := t.(type) {
	case *ast.Ident:
		if ast.IsExported(t.Name) {
			return protoFieldType{typeName: t.Name, nonNullable: true}
		}
		switch t.Name {
		case "int32", "int64", "uint32", "uint64", "string", "bool":
			return protoFieldType{typeName: t.Name}
		case "float32":
			return protoFieldType{typeName: "float"}
		case "float64":
			return protoFieldType{typeName: "double"}
		case "int":
			return protoFieldType{typeName: "int32", customType: "int"}
		}
	case *ast.StarExpr:
		pt := equivProtoType(t.X)
		pt.nonNullable = false
		if ast.IsExported(pt.typeName) {
			// only non-primitive types can be optional
			pt.optional = true
		}
		return pt
	case *ast.ArrayType:
		switch astString(t) {
		case "[]byte":
			return protoFieldType{typeName: "bytes"}
		}
		pt := equivProtoType(t.Elt)
		pt.repeated = true
		pt.optional = false // redundant
		return pt
	}
	if typ, ok := customTypeMapping[astString(t)]; ok {
		return typ
	}
	return protoFieldType{typeName: fmt.Sprintf("UNKNOWN /* add entry for %q to customTypeMapping section */", astString(t))}
}

func astString(x ast.Expr) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}

func embeddedTypeName(x ast.Expr) string {
	switch x := x.(type) {
	case *ast.StarExpr:
		return embeddedTypeName(x.X)
	case *ast.SelectorExpr:
		return x.Sel.Name
	case *ast.Ident:
		return x.Name
	default:
		log.Fatalf("embeddedTypeName: unexpected ast.Expr %T: %v", x, x)
		panic("unreachable")
	}
}

// nounForType synthesizes a reasonable argument/result name for
// something with type x.
func nounForType(x ast.Expr) string {
	switch x := x.(type) {
	case *ast.StarExpr:
		return nounForType(x.X)
	case *ast.SelectorExpr:
		return nounForType(x.Sel)
	case *ast.Ident:
		return x.Name
	case *ast.ArrayType:
		return pluralize(nounForType(x.Elt))
	default:
		log.Fatalf("embeddedTypeName: unexpected ast.Expr %T: %v", x, x)
		panic("unreachable")
	}
}

func pluralize(noun string) string {
	// quick sub-optimal hack
	esSuffixes := []string{"ch"}
	for _, suff := range esSuffixes {
		if strings.HasSuffix(noun, suff) {
			return noun + "es"
		}
	}
	if strings.HasSuffix(noun, "y") {
		return noun[:len(noun)-1] + "ies"
	}
	return noun + "s"
}

// stripServiceRelatedSuffix removes "Server" or "Service" suffixes
// from name. It is used to get a shorter, simpler name stem for
// creating synthesized names for things related to this service.
func stripServiceRelatedSuffix(name string) string {
	return strings.TrimSuffix(strings.TrimSuffix(name, "Server"), "Service")
}

func docToText(w io.Writer, text, indent string) {
	// Fix up blank lines without comments between paragraphs (these
	// slice up the doc comment in the generated Go code).
	var buf bytes.Buffer
	doc.ToText(&buf, text, indent, "", docWrap)
	b := bytes.Replace(buf.Bytes(), []byte("\n\n"), []byte("\n"+indent+"\n"), -1)
	w.Write(b)
}
