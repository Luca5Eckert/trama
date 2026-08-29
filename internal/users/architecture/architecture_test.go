package architecture

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestDependencyRule(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve current test file")
	}
	usersRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), ".."))

	forbidden := map[string][]string{
		"application":  {"/internal/users/infrastructure", "/internal/users/presentation"},
		"domain":       {"/internal/users/application", "/internal/users/infrastructure", "/internal/users/presentation"},
		"presentation": {"/internal/users/infrastructure"},
	}

	err := filepath.WalkDir(usersRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || filepath.Base(path) == "module.go" {
			return nil
		}

		layer := layerFor(usersRoot, path)
		blockedImports, protected := forbidden[layer]
		if !protected {
			return nil
		}

		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Errorf("parse %s: %v", path, err)
			return nil
		}

		for _, imported := range parsed.Imports {
			importPath, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				t.Errorf("unquote import in %s: %v", path, err)
				continue
			}
			for _, blocked := range blockedImports {
				if strings.Contains(importPath, blocked) {
					t.Errorf("%s layer imports forbidden package %q in %s", layer, importPath, path)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk users module: %v", err)
	}
}

func layerFor(root, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return ""
	}
	parts := strings.Split(filepath.ToSlash(relative), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
