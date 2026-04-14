/*
Copyright © 2023 Tech Engine
*/
package cli

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

const minGoVersion = "1.22"

//go:embed templates/*
var templatesFS embed.FS

var templateDir = filepath.Join(filepath.Dir("."), "/templates")

// startprojectCmd represents the startproject command
var startprojectCmd = &cobra.Command{
	Use:   "startproject [projectname]",
	Short: "Creates a new GoScrapy project with the specified name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var projectName = strings.TrimSpace(args[0])

		if projectName == "" {
			fmt.Printf("⚠️ please provide a projectname")
			return
		}

		templateFiles, err := fs.Glob(templatesFS, "templates/*.tmpl")

		if err != nil {
			fmt.Printf("❌ Error finding template: %v", err)
			return
		}

		if err := checkGoToolchain(); err != nil {
			fmt.Printf("❌ Error: Go toolchain not found in PATH. Please install Go (>= %s) to continue.\n", minGoVersion)
			return
		}

		fmt.Printf("\n🚀 GoScrapy generating project files. Please wait!\n\n")

		// Handle go mod init if go.mod doesn't exist
		if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
			fmt.Printf("📦 Initializing Go module: %s...\n", projectName)
			if err := runGoCommand("mod", "init", projectName); err != nil {
				fmt.Printf("⚠️  Warning: Failed to initialize go module: %v\n", err)
			}
		} else {
			// go.mod exists, check version
			content, err := os.ReadFile("go.mod")
			if err == nil {
				version := getGoVersionFromMod(string(content))
				if version != "" && !isSupportedGoVersion(version, minGoVersion) {
					fmt.Printf("⚠️  Warning: Existing go.mod uses Go %s, which is below the minimum required %s. Skipping dependency resolution.\n", version, minGoVersion)
					// We'll still generate files but won't run tidy
					generateFiles(projectName, templateFiles)
					fmt.Printf("\n✨ Congrats, %s created successfully (with warnings).\n", projectName)
					return
				}
			}
		}

		generateFiles(projectName, templateFiles)

		// ask for confirmation before resolving dependencies
		if confirm("\n📦 Do you want to resolve dependencies now (go mod tidy)?", true) {
			fmt.Printf("📦 Resolving dependencies...\n")
			if err := runGoCommand("mod", "tidy"); err != nil {
				fmt.Printf("⚠️  Warning: Failed to resolve dependencies: %v\n", err)
				fmt.Printf("You may need to run 'go mod tidy' manually.\n")
			}
		} else {
			fmt.Println("⏭️  Skipping dependency resolution. You can run 'go mod tidy' manually later.")
		}

		fmt.Printf("\n✨ Congrats, %s created successfully.", projectName)
	},
}

func generateFiles(projectName string, templateFiles []string) {
	// create [projectName] dir where we will put spider code & pipelines
	err := createDirIfNotExist(projectName)

	if err != nil {
		fmt.Printf("❌ Error creating dir '%s', %v\n", projectName, err)
		return
	}

	// create [projectName]/pipelines dir
	err = createDirIfNotExist(path.Join(projectName, "pipelines"))

	if err != nil {
		fmt.Printf("❌ Error creating dir %s/pipelines, %v\n", projectName, err)
		return
	}

	var sourceFilename string

	// Parse and execute each template
	for _, templateFile := range templateFiles {

		if templateFile == "templates/pipeline.tmpl" {
			continue
		}

		tmplContent, err := templatesFS.ReadFile(templateFile)

		if err != nil {
			fmt.Printf("❌  Error reading template: %v\n", err)
			return
		}

		tmplName := filepath.Base(templateFile)
		tmpl, err := template.New(tmplName).Parse(string(tmplContent))

		if err != nil {
			fmt.Printf("❌  Error parsing template: %v\n", err)
			return
		}

		buffer := &bytes.Buffer{}

		err = tmpl.Execute(buffer, projectName)

		if err != nil {
			fmt.Printf("❌  Error executing template: '%s', %v\n", tmpl.Name(), err)
			return
		}

		formattedCode, err := format.Source(buffer.Bytes())

		if err != nil {
			fmt.Printf("❌  Error formatting sourcecode '%s', %v\n", tmpl.Name(), err)
			return
		}

		filename := strings.TrimSuffix(tmpl.Name(), ".tmpl") + ".go"

		if templateFile == "templates/main.tmpl" {
			sourceFilename = filename
		} else {
			sourceFilename = filepath.Join(projectName, filename)
		}

		err = writeToFile(sourceFilename, formattedCode)

		if err != nil {
			fmt.Printf("❌  Error creating %s.\n", sourceFilename)
			return
		}

		fmt.Printf("✔️  %s\n", sourceFilename)
	}
}

func init() {
	rootCmd.AddCommand(startprojectCmd)
}

func writeToFile(filename string, data []byte) error {

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

func createDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		// Directory exists, prompt user for confirmation
		if !confirm(fmt.Sprintf("Directory '%s' already exists. Continue?", dir), true) {
			return nil
		}
	}

	return os.MkdirAll(dir, os.ModePerm)
}
func runGoCommand(args ...string) error {
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func confirm(prompt string, defaultValue bool) bool {
	choices := "Y/n"
	if !defaultValue {
		choices = "y/N"
	}

	fmt.Printf("%s [%s]: ", prompt, choices)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultValue
	}
	return input == "y" || input == "yes"
}

func checkGoToolchain() error {
	if _, err := exec.LookPath("go"); err != nil {
		return err
	}

	// Verify the version of the installed binary
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return fmt.Errorf("failed to check go version: %v", err)
	}

	// Output format is typically: "go version go1.22.1 windows/amd64"
	fields := strings.Fields(string(out))
	if len(fields) < 3 || !strings.HasPrefix(fields[2], "go") {
		return fmt.Errorf("unexpected 'go version' output: %s", string(out))
	}

	version := strings.TrimPrefix(fields[2], "go")
	if !isSupportedGoVersion(version, minGoVersion) {
		return fmt.Errorf("installed Go version %s is below the minimum required %s", version, minGoVersion)
	}

	return nil
}

func getGoVersionFromMod(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go ") {
			return strings.TrimPrefix(line, "go ")
		}
	}
	return ""
}

func isSupportedGoVersion(current, min string) bool {
	c, m := strings.Split(current, "."), strings.Split(min, ".")
	for i := 0; i < len(c) && i < len(m); i++ {
		cv, _ := strconv.Atoi(c[i])
		mv, _ := strconv.Atoi(m[i])
		if cv != mv {
			return cv > mv
		}
	}
	return len(c) >= len(m)
}
