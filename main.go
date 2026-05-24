package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	. "github.com/dave/jennifer/jen"
	"golang.org/x/tools/go/packages"
)

type Arguments struct {
	Type string `arg:"-t"`
}

func main() {
	ctx := context.Background()

	var args Arguments

	arg.MustParse(&args)
	if len(args.Type) == 0 {
		log.Printf("empty type provided")
		return
	}

	contents, err := generateGetters(ctx, args)
	if err != nil {
		log.Fatalf("error generating getters: %s", err.Error())
		return
	}

	if err := contents.Save(strings.ToLower(args.Type) + "_getters.go"); err != nil {
		log.Fatalf("error saving getters: %s", err.Error())
		return
	}
}

func generateGetters(ctx context.Context, args Arguments) (*File, error) {
	st, err := getStructAndPackageName(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("error reading package: %s", err.Error())
	}

	if st.s == nil {
		return nil, fmt.Errorf("struct not found for %s", args.Type)
	}

	f := NewFile(st.packageName)

	for name, path := range st.imports {
		f.ImportName(path, name)
	}

	for _, field := range st.s.Fields.List {
		if field == nil || len(field.Names) == 0 {
			continue
		}

		name := field.Names[0]
		qual, tName := getExprName(field.Type)
		if tName == "" || len(name.Name) == 0 {
			continue
		}

		if q, ok := st.imports[qual]; ok {
			qual = q
		}

		// If it's already upper no need. Unicode not supported.
		if strings.ToUpper(string(name.Name[0])) == string(name.Name[0]) {
			continue
		}

		recieverName := strings.ToLower(string(args.Type[0]))
		titleName := strings.ToTitle(string(name.Name[0])) + string(name.Name[1:])

		if qual == "" {
			f.Func().Params(
				Id(recieverName).Op("*").Id(args.Type),
			).Id(titleName).Params().Id(tName).
				Block(Return(Id(fmt.Sprintf("%s.%s", recieverName, name.Name))))
		} else {
			f.Func().Params(
				Id(recieverName).Op("*").Id(args.Type),
			).Id(titleName).Params().Qual(qual, tName).
				Block(Return(Id(fmt.Sprintf("%s.%s", recieverName, name.Name))))
		}

	}

	return f, nil
}

func getExprName(exp ast.Expr) (qual string, name string) {
	switch x := exp.(type) {
	case *ast.Ident:
		return "", x.Name
	case *ast.SelectorExpr:
		_, qual := getExprName(x.X)
		return qual, x.Sel.String()
	default:
		return "", ""
	}
}

// strct is a wrapper over *ast.StructType that includes additional data about
// the package it is in and the things imported in it's file location.
type strct struct {
	packageName string
	s           *ast.StructType
	imports     map[string]string
}

func getStructAndPackageName(
	ctx context.Context,
	args Arguments,
) (strct, error) {
	conf := packages.Config{
		Context: ctx,
	}

	pkgs, err := packages.Load(&conf)
	if err != nil {
		return strct{}, fmt.Errorf("could not load package: %s", err.Error())
	}

	if len(pkgs) == 0 {
		return strct{}, fmt.Errorf("no pacakge found: %w", err)
	}

	var st *ast.StructType
	var packageName string
	imports := make(map[string]string)

	for _, pkg := range pkgs {
		for _, fileName := range pkg.GoFiles {
			src, err := os.ReadFile(fileName)
			if err != nil {
				return strct{}, fmt.Errorf("couldn't open file: %w", err)
			}

			fset := token.NewFileSet() // positions are relative to fset

			f, err := parser.ParseFile(fset, "src.go", src, 0)
			if err != nil {
				return strct{}, fmt.Errorf("could not parse file %s: %w", fileName, err)
			}

			fileImports := make(map[string]string)

			var typeFound bool

			// Inspect the AST and print all identifiers and literals.
			ast.Inspect(f, func(n ast.Node) bool {
				var s string
				switch x := n.(type) {
				case *ast.ImportSpec:
					pathVal := strings.ReplaceAll(x.Path.Value, "\"", "")
					if x.Name != nil {
						fileImports[x.Name.Name] = pathVal
					} else {
						splitPath := strings.Split(pathVal, "/")
						if len(splitPath) > 0 {
							fileImports[splitPath[len(splitPath)-1]] = pathVal
						}
					}
				case *ast.Ident:
					s = x.Name
				case *ast.StructType:
					if typeFound {
						st = x
						packageName = pkg.Name
						imports = fileImports

						return true
					}
					s = string(src[x.Struct-1:])
				}

				typeFound = typeFound || s == args.Type

				return true
			})

			if typeFound {
				return strct{
					s:           st,
					packageName: packageName,
					imports:     imports,
				}, nil
			}
		}
	}

	return strct{
		s:           st,
		packageName: packageName,
		imports:     imports,
	}, fmt.Errorf("could not find the struct in this package")
}

// used for testing
type T struct {
	arrType   ast.ArrayType
	id        string
	num       int
	Something string
	createdAt time.Time
}
