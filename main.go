package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type DependencyMap map[string][]string

func findGoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func parseImports(filename string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	imports := []string{}
	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		imports = append(imports, path)
	}
	return imports, nil
}

func buildImportToFileMap(files []string) map[string]string {
	importToFile := make(map[string]string)
	for _, file := range files {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, file, nil, parser.PackageClauseOnly)
		if err != nil {
			continue
		}
		importToFile[f.Name.Name] = file
	}
	return importToFile
}

func detectCycles(depMap DependencyMap) map[string]map[string]bool {
	cycles := make(map[string]map[string]bool)
	var visit func(string, map[string]bool)
	visit = func(file string, path map[string]bool) {
		if path[file] {
			if cycles[file] == nil {
				cycles[file] = make(map[string]bool)
			}
			for k := range path {
				cycles[file][k] = true
			}
			return
		}
		path[file] = true
		for _, dep := range depMap[file] {
			visit(dep, path)
		}
		delete(path, file)
	}
	for file := range depMap {
		visit(file, make(map[string]bool))
	}
	return cycles
}

func main() {
	root := "."
	files, err := findGoFiles(root)
	if err != nil {
		fmt.Println("Ошибка при поиске файлов:", err)
		return
	}

	depMap := make(DependencyMap)
	for _, file := range files {
		imports, err := parseImports(file)
		if err != nil {
			fmt.Printf("Ошибка при парсинге %s: %v\n", file, err)
			continue
		}
		depMap[file] = imports
	}

	importToFile := buildImportToFileMap(files)

	fileToPkg := make(map[string]string)
	for pkg, file := range importToFile {
		fileToPkg[file] = pkg
	}

	importPathToFile := make(map[string]string)
	for _, file := range files {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, file, nil, parser.PackageClauseOnly)
		if err != nil {
			continue
		}
		importPathToFile[f.Name.Name] = file
	}

	cycles := detectCycles(depMap)

	fmt.Println("Зависимости между файлами:")
	for file, imports := range depMap {
		fmt.Printf("%s:\n", file)
		seen := make(map[string]bool)
		for _, imp := range imports {
			mark := ""
			if _, ok := importPathToFile[imp]; !ok {
				mark = " ❌ (файл не найден)"
			}
			if seen[imp] {
				mark = " ❌ (дублируется)"
			}
			seen[imp] = true
			if cycles[file] != nil && cycles[file][imp] {
				mark = " ❌ (цикл)"
			}
			fmt.Printf("  └── %s%s\n", imp, mark)
		}
	}
}
