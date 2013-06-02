package main

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func init() {
	importPaths = make(map[string]string)
}

var importPaths map[string]string
var rootPackage string

func getImportPaths(folder, packageName string) {
	if _, found := importPaths[folder]; found {
		return
	}
	importPaths[folder] = packageName

	info, err := os.Stat(folder)
	if err != nil {
		return
	}

	if info.IsDir() {
		f, _ := os.Open(folder)
		defer f.Close()
		dn, _ := f.Readdirnames(-1)
		for _, filename := range dn {
			if filename[0] == '.' {
				continue
			}
			filename = filepath.Join(folder, filename)
			if strings.HasSuffix(filename, ".go") {
				fset := token.NewFileSet()
				p, _ := parser.ParseFile(fset, filename, nil, parser.ImportsOnly)
				for _, s := range p.Imports {
					gopath := os.Getenv("GOPATH")
					importPackage, _ := strconv.Unquote(s.Path.Value)
					importPath := filepath.Join(gopath, "src", importPackage)
					if _, statErr := os.Stat(importPath); statErr == nil {
						getImportPaths(importPath, importPackage)
					}
				}
			}
		}
	}
	return
}

func WriteCombinedLicense(w io.Writer) {
	for path, pkg := range importPaths {
		f, _ := os.Open(path)
		dn, _ := f.Readdirnames(-1)
		foundLicense := false
		for _, filename := range dn {
			if strings.HasPrefix(strings.ToUpper(filename), "LICENSE") {
				foundLicense = true
				filename = filepath.Join(path, filename)
				lfile, _ := os.Open(filename)
				bufReader := bufio.NewReader(lfile)
				defer lfile.Close()
				if pkg != rootPackage {
					fmt.Fprintln(w, pkg)
					fmt.Fprintln(w, "----------")
				}
				bufReader.WriteTo(w)
				fmt.Fprintln(w, "")
				break
			}
		}
		if !foundLicense && pkg != rootPackage {
			fmt.Fprintln(w, "")
			fmt.Fprintf(w, "%s: no license", pkg)
			fmt.Fprintln(w, "")
		}
	}
}

func main() {
	// Print the imports from the file's AST.
	pwd, _ := os.Getwd()
	rootPackage = filepath.Base(pwd)
	getImportPaths(pwd, rootPackage)
	WriteCombinedLicense(os.Stdout)
}
