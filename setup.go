//go:build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	r := bufio.NewReader(os.Stdin)

	project := prompt(r, "Repository name: ")
	owner := prompt(r, "GitHub owner (username or org): ")
	npmScope := promptDefault(r, fmt.Sprintf("npm scope [%s]: ", owner), owner)
	pypiName := promptDefault(r, fmt.Sprintf("PyPI package name [%s-cli]: ", project), project+"-cli")
	brewTap := promptDefault(r, fmt.Sprintf("Homebrew tap [%s/tap]: ", owner), owner+"/tap")

	fmt.Printf("\nConfiguring:\n")
	fmt.Printf("  Project:  %s\n", project)
	fmt.Printf("  Owner:    %s\n", owner)
	fmt.Printf("  npm:      @%s/%s\n", npmScope, project)
	fmt.Printf("  PyPI:     %s\n", pypiName)
	fmt.Printf("  Brew tap: %s\n\n", brewTap)

	files := gitTrackedFiles()

	replacements := []struct{ old, new string }{
		{"YOUR_OWNER/tap", brewTap},
		{"YOUR_OWNER", owner},
		{"YOUR_PROJECT-cli", pypiName},
		{"YOUR_PROJECT_cli", project + "_cli"},
		{"YOUR_PROJECT", project},
	}

	for _, f := range files {
		replaceInFile(f, replacements)
	}

	// Rename python package directory
	pyOld := filepath.Join("python-package", "YOUR_PROJECT_cli")
	pyNew := filepath.Join("python-package", project+"_cli")
	if _, err := os.Stat(pyOld); err == nil {
		os.Rename(pyOld, pyNew)
	}

	// Rename npm runner script
	jsOld := filepath.Join("npm-package", "bin", "YOUR_PROJECT.js")
	jsNew := filepath.Join("npm-package", "bin", project+".js")
	if _, err := os.Stat(jsOld); err == nil {
		os.Rename(jsOld, jsNew)
	}

	// Remove this setup file
	os.Remove("setup.go")

	fmt.Println("Done. Next steps:")
	fmt.Printf("  1. Set up secrets: HOMEBREW_TAP_GITHUB_TOKEN, npm, PyPI trusted publishing\n")
	fmt.Printf("  2. Create your homebrew tap repo: github.com/%s\n", brewTap)
	fmt.Printf("  3. Build your CLI in main.go\n")
	fmt.Printf("  4. Push a v* tag to release\n")
}

func prompt(r *bufio.Reader, label string) string {
	for {
		fmt.Print(label)
		s, _ := r.ReadString('\n')
		s = strings.TrimSpace(s)
		if s != "" {
			return s
		}
	}
}

func promptDefault(r *bufio.Reader, label, def string) string {
	fmt.Print(label)
	s, _ := r.ReadString('\n')
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	return s
}

func gitTrackedFiles() []string {
	out, err := exec.Command("git", "ls-files").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "git ls-files failed:", err)
		os.Exit(1)
	}
	var files []string
	for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if f != "" {
			files = append(files, f)
		}
	}
	return files
}

func replaceInFile(path string, replacements []struct{ old, new string }) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	content := string(data)
	original := content
	for _, r := range replacements {
		content = strings.ReplaceAll(content, r.old, r.new)
	}
	if content != original {
		os.WriteFile(path, []byte(content), 0)
	}
}
