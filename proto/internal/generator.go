// +build ignore

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Parse parses the source code of a single Go source file and returns the
// corresponding structs.
func Parse(r io.Reader, name string) ([]Struct, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, data, 0)
	if err != nil {
		return nil, err
	}

	var structs []Struct
	for _, decl := range f.Decls {
		gendecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if len(gendecl.Specs) < 1 {
			continue
		}
		typespec, ok := gendecl.Specs[0].(*ast.TypeSpec)
		if !ok {
			continue
		}

		structname := typespec.Name.Name
		switch node := typespec.Type.(type) {
		case *ast.StructType:
			s := parseFields(structname, node)
			structs = append(structs, s)
		case *ast.Ident:
			structs = append(structs, Struct{
				Name: structname,
				Type: node.Name,
			})
		}
	}
	return structs, err
}

// Field represents a go struct field.
type Field struct {
	Name string
	Type string
}

// Struct represents a go struct / type declaration.
type Struct struct {
	Name   string
	Type   string
	Fields []Field
}

func parseName(field *ast.Field, fieldtype string) string {
	if len(field.Names) == 0 {
		return fieldtype
	}
	return field.Names[0].Name
}

func parseFields(name string, node *ast.StructType) Struct {
	s := Struct{Name: name, Type: "struct"}
	for _, field := range node.Fields.List {
		fieldtype := ""
		switch t := field.Type.(type) {
		case *ast.ArrayType:
			fieldtype = t.Elt.(*ast.Ident).Name
		case *ast.Ident:
			fieldtype = t.Name
		case *ast.SelectorExpr:
			fieldtype = fmt.Sprintf("%s.%s", t.X, t.Sel)
		default:
			continue
		}

		s.Fields = append(s.Fields, Field{
			Name: parseName(field, fieldtype),
			//Name: field.Names[0].Name,
			Type: fieldtype,
		})
	}
	return s
}

type Generator struct {
	buf     *bytes.Buffer
	pkgname string
}

func NewGenerator(pkgname string) *Generator {
	return &Generator{buf: bytes.NewBuffer(nil), pkgname: pkgname}
}

func (g *Generator) Bytes() []byte { return g.buf.Bytes() }

func (g *Generator) printf(format string, args ...interface{}) {
	fmt.Fprintf(g.buf, format, args...)
}

func (g *Generator) print(format string, args ...interface{}) {
	fmt.Fprintf(g.buf, format+"\n", args...)
}

func (g *Generator) genStructReset(s Struct) {
	g.print("// Reset resets all state.")
	g.print("func (m *%s) Reset() { *m = %s{} }", s.Name, s.Name)
}

func (g *Generator) genStructType(s Struct) {
	g.print("// MessageType returns the message type.")
	g.print("func (m %s) MessageType() MessageType { return Message%s }",
		s.Name, s.Name,
	)
}

func (g *Generator) genStructLen(s Struct) {
	g.print("// Len returns the length of the message in bytes.")
	g.printf("func (m %s) Len() int {", s.Name)
	if len(s.Fields) == 0 {
		g.printf("return 0 }\n\n")
		return
	}
	g.printf("return ")
	for i, field := range s.Fields {
		switch field.Type {
		case "string":
			g.printf("2 + len(m.%s)", field.Name)
		case "uint64":
			g.printf("8")
		case "uint32", "Mode", "Flag":
			g.printf("4")
		case "uint16":
			g.printf("2")
		case "uint8", "MessageType":
			g.printf("1")
		case "Qid":
			g.printf("13")
		case "unix.Timespec":
			g.printf("16")
		}
		if i < len(s.Fields)-1 {
			g.printf("+")
		}
	}
	g.printf("}\n\n")
}

func (g *Generator) genStructDecode(s Struct) {
	g.print("// Decode decodes from the given binary.Buffer.")
	if len(s.Fields) == 0 {
		g.print("func (m *%s) Decode(buf *binary.Buffer) {}", s.Name)
		return
	}
	g.print("func (m *%s) Decode(buf *binary.Buffer) {", s.Name)
	for _, field := range s.Fields {
		switch field.Type {
		case "string":
			g.print("m.%s = buf.String()", field.Name)
		case "uint64":
			g.print("m.%s = buf.Uint64()", field.Name)
		case "uint32":
			g.print("m.%s = buf.Uint32()", field.Name)
		case "uint16":
			g.print("m.%s = buf.Uint16()", field.Name)
		case "uint8":
			g.print("m.%s = buf.Uint8()", field.Name)
		case "Mode":
			g.print("m.%s = Mode(buf.Uint32())", field.Name)
		case "Flag":
			g.print("m.%s = Flag(buf.Uint32())", field.Name)
		case "MessageType":
			g.print("m.%s = MessageType(buf.Uint8())", field.Name)
		case "Qid":
			g.print("m.%s.Decode(buf)", field.Name)
		case "unix.Timespec":
			g.print("m.%s = decodeTimespec(buf)", field.Name)
		}
	}
	g.print("}\n")
}

func (g *Generator) genStructEncode(s Struct) {
	g.print("// Encode encodes to the given binary.Buffer.")
	if len(s.Fields) == 0 {
		g.print("func (m %s) Encode(buf *binary.Buffer) {}", s.Name)
		return
	}
	g.print("func (m %s) Encode(buf *binary.Buffer) {", s.Name)
	for _, field := range s.Fields {
		switch field.Type {
		case "string":
			g.print("buf.PutString(m.%s)", field.Name)
		case "uint64":
			g.print("buf.PutUint64(m.%s)", field.Name)
		case "uint32":
			g.print("buf.PutUint32(m.%s)", field.Name)
		case "uint16":
			g.print("buf.PutUint16(m.%s)", field.Name)
		case "uint8":
			g.print("buf.PutUint8(m.%s)", field.Name)
		case "Mode":
			g.print("buf.PutUint32(uint32(m.%s))", field.Name)
		case "Flag":
			g.print("buf.PutUint32(uint32(m.%s))", field.Name)
		case "MessageType":
			g.print("buf.PutUint8(uint8(m.%s))", field.Name)
		case "Qid":
			g.print("m.%s.Encode(buf)", field.Name)
		case "unix.Timespec":
			g.print("encodeTimespec(buf, m.%s)", field.Name)
		}
	}
	g.print("}\n")
}

func toSnake(s string) string {
	n := ""
	for i, v := range s {
		isChanged := false
		if i+1 < len(s) {
			next := s[i+1]
			if (v >= 'A' && v <= 'Z' && next >= 'a' && next <= 'z') ||
				(v >= 'a' && v <= 'z' && next >= 'A' && next <= 'Z') {
				isChanged = true
			}
		}

		if i > 0 && n[len(n)-1] != '_' && isChanged {
			if v >= 'A' && v <= 'Z' {
				n += string('_') + string(v)
			} else if v >= 'a' && v <= 'z' {
				n += string(v) + string('_')
			}
		} else if v == ' ' || v == '_' || v == '-' {
			n += string('_')
		} else {
			n += string(v)
		}
	}

	return strings.ToLower(n)
}

func (g *Generator) genStructString(s Struct) {
	g.print("// String implements fmt.Stringer.")
	if len(s.Fields) == 0 {
		g.print("func (m %s) String() string { return \"\" }", s.Name)
		return
	}

	g.printf("func (m %s) String() string { return fmt.Sprintf(", s.Name)
	str := `"`
	for _, field := range s.Fields {
		switch field.Type {
		case "string":
			str += fmt.Sprintf("%s:%s ", toSnake(field.Name), "%%q")
		case "uint64", "uint32", "uint16", "uint8", "Flag":
			str += fmt.Sprintf("%s:%s ", toSnake(field.Name), "%%d")
		case "Mode":
			str += fmt.Sprintf("%s:%s ", toSnake(field.Name), "%%q")
		case "MessageType":
			str += fmt.Sprintf("%s ", "%%s")
		case "Qid":
			str += fmt.Sprintf("%s ", "%%s")
		case "unix.Timespec":
			str += fmt.Sprintf("%s:%s ", toSnake(field.Name), "%%d")
		}
	}
	if len(str) > 0 {
		str = str[:len(str)-1]
	}
	str += `", `
	for _, field := range s.Fields {
		if field.Type == "unix.Timespec" {
			str += fmt.Sprintf("m.%s.Nano(), ", field.Name)
			continue
		}
		str += fmt.Sprintf("m.%s, ", field.Name)
	}
	if len(str) > 1 {
		str = str[:len(str)-2]
	}

	g.printf(str)
	g.print(")}\n")
}

func (g *Generator) genStructTest(s Struct) {
	g.print("{&%s{}, &%s{}},", s.Name, s.Name)
	if len(s.Fields) == 0 {
		return
	}

	g.printf("{&%s{", s.Name)
	for _, field := range s.Fields {
		switch field.Type {
		case "string":
			g.printf("%s: string16.String(),", field.Name)
		case "uint64":
			g.printf("%s: math.MaxUint64,", field.Name)
		case "uint32", "Mode", "Flag":
			g.printf("%s: math.MaxUint32,", field.Name)
		case "uint16":
			g.printf("%s: math.MaxUint16,", field.Name)
		case "uint8":
			g.printf("%s: math.MaxUint8,", field.Name)
		case "MessageType":
			g.printf("%s: math.MaxUint8,", field.Name)
		case "Qid":
			g.printf("%s: Qid{Type: math.MaxUint8, Version: math.MaxUint32, Path: math.MaxUint64},", field.Name)
		}
	}
	g.print("}, &%s{}},", s.Name)
}

var ignore = map[string]bool{
	"Twalk":    true,
	"Rwalk":    true,
	"Rread":    true,
	"Twrite":   true,
	"Rreaddir": true,
}

func (g *Generator) Generate(r io.Reader, name string) error {
	g.print("// THIS FILE IS AUTOMATICALLY GENERATED by `go run internal/generator.go`.")
	g.print("// EDIT %s INSTEAD.\n", name)

	g.print("package %s\n", g.pkgname)
	g.print(`import (
		"fmt"

		"github.com/azmodb/ninep/binary"
	)` + "\n")

	structs, err := Parse(r, name)
	if err != nil {
		return err
	}

	for _, s := range structs {
		if _, found := ignore[s.Name]; found {
			continue
		}
		if s.Type == "struct" {
			g.genStructType(s)
			g.genStructString(s)
			g.genStructLen(s)
			g.genStructReset(s)
			g.genStructEncode(s)
			g.genStructDecode(s)
		}
	}

	return nil
}

func (g *Generator) GeneratePool(w io.Writer, srcs ...string) error {
	g.print("// THIS FILE IS AUTOMATICALLY GENERATED by `go run internal/generator.go`.")
	g.print("// EDIT legacy.go, linux.go INSTEAD.\n")

	g.print("package %s\n", g.pkgname)
	g.print(`import "sync"` + "\n")

	buf := &bytes.Buffer{}
	for _, src := range srcs {
		r, err := os.Open(src)
		if err != nil {
			return err
		}
		defer r.Close()

		structs, err := Parse(r, "<pool.go>")
		if err != nil {
			return err
		}

		for _, s := range structs {
			g.print("var %sPool = sync.Pool{", strings.ToLower(s.Name))
			g.print("	New: func() interface{} { return &%s{} },", s.Name)
			g.print("}\n")

			g.print("// Alloc%s selects an arbitrary item from the %s Pool", s.Name, s.Name)
			g.print("// removes it, and returns it to the caller.")
			g.print("func Alloc%s() *%s {", s.Name, s.Name)
			g.print("	return %sPool.Get().(*%s)", strings.ToLower(s.Name), s.Name)
			g.print("}\n")

			g.print("// Release resets all state and adds m to the %s pool.", s.Name)
			g.print("func (m *%s) Release() {", s.Name)
			g.print("	m.Reset()")
			g.print("	%sPool.Put(m)", strings.ToLower(s.Name))
			g.print("}\n")
		}

		if _, err = io.Copy(buf, g.buf); err != nil {
			return err
		}
	}

	data, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}
	_, err = w.Write(data)
	return err
}

func (g *Generator) GenerateTest(r io.Reader, name string) error {
	g.print("// THIS FILE IS AUTOMATICALLY GENERATED by `go run internal/generator.go`.")
	g.print("// EDIT %s INSTEAD.\n", name)

	g.print("package %s\n", g.pkgname)
	g.print(`import "math"` + "\n")

	structs, err := Parse(r, name)
	if err != nil {
		return err
	}

	name = name[:len(name)-3] // TODO
	name = strings.Title(name)

	g.print("var generated%sPackets = []packet {", name)
	for _, s := range structs {
		if _, found := ignore[s.Name]; found {
			continue
		}
		switch s.Type {
		case "uint16", "uint32":
		case "struct":
			g.genStructTest(s)
		}
	}
	g.print("}")
	return nil
}

type op func(io.Reader, string) ([]byte, error)

func source(g *Generator) op {
	return func(r io.Reader, name string) ([]byte, error) {
		if err := g.Generate(r, name); err != nil {
			return nil, err
		}
		data := g.Bytes()
		g.buf.Reset()
		return data, nil
	}
}

func test(g *Generator) op {
	return func(r io.Reader, name string) ([]byte, error) {
		if err := g.GenerateTest(r, name); err != nil {
			return nil, err
		}
		data := g.Bytes()
		g.buf.Reset()
		return data, nil
	}
}

func generate(dst, src string, fn op) error {
	r, err := os.Open(src)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	data, err := fn(r, filepath.Base(src))
	if err != nil {
		return err
	}

	data, err = format.Source(data)
	if err != nil {
		panic(err)
	}

	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = w.Write(data)
	return err
}

func genPool(g *Generator, dst string, srcs ...string) error {
	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer w.Close()

	return g.GeneratePool(w, srcs...)
}

func main() {
	g := NewGenerator("proto")

	if err := genPool(g, "pool_gen.go", "legacy.go", "linux.go"); err != nil {
		panic(err)
	}

	if err := generate("legacy_test.go", "legacy.go", test(g)); err != nil {
		panic(err)
	}
	if err := generate("linux_test.go", "linux.go", test(g)); err != nil {
		panic(err)
	}

	if err := generate("legacy_gen.go", "legacy.go", source(g)); err != nil {
		panic(err)
	}
	if err := generate("linux_gen.go", "linux.go", source(g)); err != nil {
		panic(err)
	}
}
