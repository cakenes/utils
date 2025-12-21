package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var css = make(map[string]map[string]bool)
var dynamic = make(map[string]bool) // Tracks if dynamic usage detected for a CSS file

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var dir string
	if len(os.Args) >= 2 {
		dir = os.Args[1]
	} else {
		fmt.Println("No directory argument provided, using current working directory.")
		dir = cwd
	}

    scanDir, err := filepath.Abs(dir)
    if err != nil {
        fmt.Printf("Error getting absolute path: %v\n", err)
        return
    }
	
	// Step 1: Find all .module.css files and extract classes
	err = filepath.WalkDir(scanDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".module.css") {
			absPath, _ := filepath.Abs(path)
			parseCSS(absPath)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking for CSS: %v\n", err)
		return
	}

	// Step 2: Find all .tsx files and check usages
	err = filepath.WalkDir(scanDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".tsx") {
			absPath, _ := filepath.Abs(path)
			checkTSX(absPath, cwd)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking for TSX: %v\n", err)
		return
	}

	// Step 3: Report unused
	for cssPath, classes := range css {
		var unused []string
		for cls, used := range classes {
			if !used {
				unused = append(unused, cls)
			}
		}

		if len(unused) > 0 {
			relPath, _ := filepath.Rel(cwd, cssPath)
			fmt.Printf("File: %s\n", relPath)
			if dynamic[cssPath] {
				fmt.Println("  (Dynamic style usage detected in one or more consumers, results may be inaccurate)")
			}
			for _, cls := range unused {
				fmt.Printf("  - %s\n", cls)
			}
			fmt.Println("")
		}
	}
}

func parseCSS(path string) {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", path, err)
		return
	}
	content := string(contentBytes)

	// Remove comments
	commentRe := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	content = commentRe.ReplaceAllString(content, "")

	// Extract .classname followed by a valid terminator
	re := regexp.MustCompile(`\.([a-zA-Z_-][a-zA-Z0-9_-]*)`)
	matches := re.FindAllStringIndex(content, -1)

	classes := make(map[string]bool)
	for _, loc := range matches {
		start, end := loc[0], loc[1]
		className := content[start+1 : end] // Skip the dot

		// Check what comes after
		if end < len(content) {
			nextChar := content[end]

			// Valid terminators for a class selector
			switch nextChar {
			case ' ', '\t', '\n', '\r', ',', '{', ':', '.', '>', '+', '~', '[':
				classes[className] = false
			}
		} else {
			classes[className] = false
		}
	}

	if len(classes) > 0 {
		css[path] = classes
	}
}

func checkTSX(path, projectRoot string) {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", path, err)
		return
	}
	content := string(contentBytes)

	// Find imports of .module.css
	importRe := regexp.MustCompile(`import\s+(\w+)\s+from\s+['"](.+\.module\.css)['"]`)
	matches := importRe.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		varName := match[1]
		importPath := match[2]

		// Resolve absolute path of css
		var cssAbsPath string

		// Handle alias or relative path, assuming @/ maps to src/
		if strings.HasPrefix(importPath, "@") {
			cleanImport := strings.TrimPrefix(importPath, "@")
            cleanImport = strings.TrimPrefix(cleanImport, "/")
			cssAbsPath = filepath.Join(projectRoot, "src", cleanImport)
		} else {
			// Relative path
			dir := filepath.Dir(path)
			cssAbsPath = filepath.Join(dir, importPath)
		}
		
		cssAbsPath = filepath.Clean(cssAbsPath)

		if classes, ok := css[cssAbsPath]; ok {
			// Check for dynamic usage: varName[]
			dynamicRe := regexp.MustCompile(regexp.QuoteMeta(varName) + `\[[^'"]`)
			if dynamicRe.MatchString(content) {
				dynamic[cssAbsPath] = true
			}

			for cls := range classes {
				// varName.cls
				pattern1 := regexp.QuoteMeta(varName) + `\.` + regexp.QuoteMeta(cls) + `\b`
				// varName['cls'] or varName["cls"]
				pattern2 := regexp.QuoteMeta(varName) + `\[['"]` + regexp.QuoteMeta(cls) + `['"]\]`

				matched1, _ := regexp.MatchString(pattern1, content)
				matched2, _ := regexp.MatchString(pattern2, content)

				if matched1 || matched2 {
					classes[cls] = true
				}
			}
		}
	}
}
