package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	ROOT_DIR        = "."
	OUTPUT_FILE     = filepath.Join("docs", "4-folder_tree.txt")
	EXCLUDE_DEFAULT = []string{}
	INCLUDE_DOTS    = []string{}
	DOTDIRS_SHALLOW = []string{".venv", ".env", "node_modules", ".git", "third_party"}
	INDENT          = 2
)

func pSprintf(f string, a ...interface{}) string {
	s := fmt.Sprintf(f, a...) /*fmt.Println(s);*/
	return s
}

func sortedEntries(path string) ([]string, error) {
	es, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var d, f []string
	for _, e := range es {
		if e.IsDir() {
			d = append(d, e.Name())
		} else {
			f = append(f, e.Name())
		}
	}
	sort.Slice(d, func(i, j int) bool { return strings.ToLower(d[i]) < strings.ToLower(d[j]) })
	sort.Slice(f, func(i, j int) bool { return strings.ToLower(f[i]) < strings.ToLower(f[j]) })
	return append(d, f...), nil
}

func shouldExclude(name string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}
	for _, p := range patterns {
		if p == name {
			return true
		}
		if ok, _ := filepath.Match(p, name); ok {
			return true
		}
	}
	return false
}

func printTree(root string, exclude []string, includeDots []string, dotShallowList []string, indent int) ([]string, error) {
	inc := map[string]struct{}{}
	for _, s := range includeDots {
		if s != "" {
			inc[s] = struct{}{}
		}
	}
	dotShallow := map[string]struct{}{}
	for _, s := range dotShallowList {
		if s != "" {
			dotShallow[s] = struct{}{}
		}
	}
	out := []string{}

	var rec func(path, prefix string, isLast bool, level int) error
	rec = func(path, prefix string, isLast bool, level int) error {
		name := filepath.Base(path)
		if level == 0 {
			if abs, err := filepath.Abs(path); err == nil {
				name = filepath.Base(abs)
			} else if name == "" {
				name = path
			}
			out = append(out, pSprintf("%s", name))
		} else {
			display := name
			suffix := "file"
			if fi, err := os.Stat(path); err == nil && fi.IsDir() {
				display += "/"
				suffix = ""
			}
			var connector string
			if isLast {
				connector = "└" + strings.Repeat("─", indent-1) + " "
				if suffix == "file" {
					suffix = "\n" + strings.TrimRight(prefix, " ")
				}
			}
			if !isLast {
				connector = "├" + strings.Repeat("─", indent-1) + " "
			}
			out = append(out, pSprintf("%s%s%s%s", prefix, connector, display, suffix))
		}

		fi, err := os.Stat(path)
		if err != nil || !fi.IsDir() {
			return nil
		}
		baseprefix := prefix
		if isLast {
			baseprefix += strings.Repeat(" ", indent)
		} else {
			baseprefix += "│" + strings.Repeat(" ", indent)
		}

		entries, err := sortedEntries(path)
		if err != nil {
			return nil
		}
		filtered := []string{}
		for _, e := range entries {
			if shouldExclude(e, exclude) {
				continue
			}
			if strings.HasPrefix(e, ".") {
				full := filepath.Join(path, e)
				if fi2, err := os.Stat(full); err == nil && fi2.IsDir() {
					filtered = append(filtered, e)
				} else {
					if _, ok := inc[e]; ok {
						filtered = append(filtered, e)
					}
				}
			} else {
				filtered = append(filtered, e)
			}
		}

		for i, e := range filtered {
			full := filepath.Join(path, e)
			last := (i == len(filtered)-1)
			if _, ok := dotShallow[e]; ok {
				if fi2, err := os.Stat(full); err == nil && fi2.IsDir() {
					con := "├" + strings.Repeat("─", indent-1) + " "
					if last {
						con = "└" + strings.Repeat("─", indent-1) + " "
					}
					out = append(out, pSprintf("%s%s%s/", baseprefix, con, e))
					continue
				}
			}
			if err := rec(full, baseprefix, last, level+1); err != nil {
				return err
			}
		}
		return nil
	}

	if err := rec(root, "", true, 0); err != nil {
		return out, err
	}
	return out, nil
}

func main() {
	out, err := printTree(ROOT_DIR, EXCLUDE_DEFAULT, INCLUDE_DOTS, DOTDIRS_SHALLOW, INDENT)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if OUTPUT_FILE != "" {
		content := strings.Join(out, "\n") + "\n"
		if err := os.WriteFile(OUTPUT_FILE, []byte(content), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "write error:", err)
			os.Exit(1)
		}
		if absPath, err := filepath.Abs(OUTPUT_FILE); err != nil {
			fmt.Printf("Output written to %s\n", OUTPUT_FILE)
		} else {
			fmt.Printf("Output written to %s\n", absPath)
		}
	}
}
