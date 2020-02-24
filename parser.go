package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ExportDirConst(w io.Writer, path string) error {
	filter := func(finf os.FileInfo) bool {
		return !strings.HasSuffix(finf.Name(), "_test.go")
	}
	fset := token.NewFileSet() // positions are relative to fset
	pkgs, err := parser.ParseDir(fset, path, filter, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(w, "err: %v\n", err)
		return err
	}
	for pkgName, pkg := range pkgs {
		files := []*ast.File{}
		for _, file := range pkg.Files {
			files = append(files, file)
		}
		if err := exportFilesConst(w, pkgName, fset, files...); err != nil {
			return err
		}
	}
	return nil
}

func ExportFilesConst(w io.Writer, pkg string, files ...string) error {
	for _, file := range files {
		if err := ExportFileConst(w, pkg, file); err != nil {
			return err
		}
	}
	return nil
}

func ExportFileConst(w io.Writer, pkg, file string) error {
	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	return exportFilesConst(w, pkg, fset, f)
}

func exportFilesConst(w io.Writer, pkgName string, fset *token.FileSet, files ...*ast.File) error {
	// do type check to evaluate the values of the constants
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check(pkgName, fset, files, nil)
	if err != nil {
		return err
	}

	for _, file := range files {
		fmt.Fprintf(w, "// package %v, file %v\n\n", pkgName, filepath.Base(fset.Position(file.Package).Filename))
		if err := exportFileConst(w, pkg, file); err != nil {
			return err
		}
	}
	return nil
}

func exportFileConst(w io.Writer, pkg *types.Package, file *ast.File) error {
	for _, decl := range file.Decls {
		g, _ := decl.(*ast.GenDecl)
		if g == nil || g.Tok != token.CONST {
			continue
		}
		if g.Doc != nil {
			fmt.Fprintf(w, "// %v", g.Doc.Text())
		}
		nameMaxLen := 0
		for _, spec := range g.Specs {
			switch s := spec.(type) {
			case *ast.ValueSpec:
				name := s.Names[0].Name
				if token.IsExported(name) && nameMaxLen < len(name) {
					nameMaxLen = len(name)
				}
			}
		}
		sfmt := fmt.Sprintf("%%-%ds = %%v", nameMaxLen)
		for _, spec := range g.Specs {
			switch s := spec.(type) {
			case *ast.ValueSpec:
				name := s.Names[0].Name
				if !token.IsExported(name) {
					continue
				}

				c, _ := pkg.Scope().Lookup(name).(*types.Const)
				if c == nil {
					continue
				}
				if s.Doc != nil {
					fmt.Fprintf(w, "// %v", s.Doc.Text())
				}
				// fsfile := fset.File(s.Pos())
				fmt.Fprintf(w, sfmt, name, c.Val().ExactString())
				if s.Comment != nil {
					// assume the comment starts on the same line as the statement ends
					// if fsfile.Line(s.End()) < fsfile.Line(s.Comment.Pos()) {
					// 	fmt.Fprintln(w)
					// } else {
					fmt.Fprint(w, " ")
					// }
					fmt.Fprintf(w, "// %v", s.Comment.Text())
				} else {
					fmt.Fprintln(w)
				}
			}
		}
		fmt.Fprintln(w)
	}
	return nil
}
