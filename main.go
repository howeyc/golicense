package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Returns a map of pkg -> path references
func getImportPaths(folder, packageName string) (importPaths map[string]string) {
	importPaths = make(map[string]string)

	var gip func(x, y string)
	gip = func(folder, packageName string) {
		if _, found := importPaths[packageName]; found {
			return
		}
		importPaths[packageName] = folder

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
						importPackage, _ := strconv.Unquote(s.Path.Value)
						gopath := os.Getenv("GOPATH")
						importPath := filepath.Join(gopath, "src", importPackage)
						goroot := os.Getenv("GOROOT")
						importRoot := filepath.Join(goroot, "src", "pkg", importPackage)
						if _, statErr := os.Stat(importPath); statErr == nil {
							gip(importPath, importPackage)
						} else if _, statErr := os.Stat(importRoot); statErr == nil {
							gip(importRoot, importPackage)
						}
					}
				}
			}
		}
	}

	gip(folder, packageName)
	return
}

// Returns a list of all License files in GOPATH and GOROOT
func getLicensePaths() []string {
	licensePaths := make([]string, 0)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasPrefix(strings.ToUpper(filepath.Base(path)), "LICENSE") {
			licensePaths = append(licensePaths, path)
		}

		return nil
	}

	filepath.Walk(os.Getenv("GOPATH"), walkFn)
	filepath.Walk(os.Getenv("GOROOT"), walkFn)

	return licensePaths
}

type License struct {
	Path    string
	Imports []string
}

// Writes all licenses
func writeCombinedLicense(w io.Writer, licenses []License) {
	for _, license := range licenses {
		if len(license.Imports) > 0 {
			if license.Path == filepath.Join(os.Getenv("GOROOT"), "LICENSE") {
				fmt.Fprintln(w, "Go Standard Library")
			} else {
				fmt.Fprintln(w, strings.Join(license.Imports, ",\n"))
			}
		} else {
			continue
		}
		fmt.Fprintln(w, "----------")
		if license.Path == "" {
			fmt.Fprintln(w, "no license found")
			fmt.Fprintln(w, "")
			continue
		}
		lfile, _ := os.Open(license.Path)
		bufReader := bufio.NewReader(lfile)
		defer lfile.Close()
		bufReader.WriteTo(w)
		fmt.Fprintln(w, "")
	}
}

// Given the imports used, find all unique licenses
func getCombinedLicenses(licensePaths []string, importPaths map[string]string) []License {
	licenses := make([]License, 1)
	for pkg, ipath := range importPaths {
		foundLicense := false
		for _, lpath := range licensePaths {
			if filepath.Dir(lpath) == ipath || strings.HasPrefix(ipath, filepath.Dir(lpath)) {
				updatedLicense := false
				for idx, _ := range licenses {
					if licenses[idx].Path == lpath {
						licenses[idx].Imports = append(licenses[idx].Imports, pkg)
						updatedLicense = true
					}
				}
				if !updatedLicense {
					licenses = append(licenses, License{lpath, []string{pkg}})
				}
				foundLicense = true
				break
			}
		}
		if !foundLicense {
			licenses[0].Imports = append(licenses[0].Imports, pkg)
		}
	}
	return licenses
}

func main() {
	flag.Parse()
	pwd, _ := os.Getwd()
	if flag.Arg(0) != "" {
		pwd = filepath.Join(os.Getenv("GOPATH"), "src", flag.Arg(0))
	}
	rootPackage := filepath.Base(pwd)
	writeCombinedLicense(os.Stdout, getCombinedLicenses(getLicensePaths(), getImportPaths(pwd, rootPackage)))
}
