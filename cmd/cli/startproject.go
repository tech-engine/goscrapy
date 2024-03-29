/*
Copyright © 2023 Tech Engine
*/
package cli

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

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

		fmt.Printf("\n🚀 GoScrapy generating project files. Please wait!\n\n")

		// create [projectName] dir where we will put spider code & pipelines
		err = createDirIfNotExist(projectName)

		if err != nil {
			fmt.Printf("❌ Error creating dir '%s', %v", projectName, err)
			return
		}

		// create [projectName]/pipelines dir
		err = createDirIfNotExist(path.Join(projectName, "pipelines"))

		if err != nil {
			fmt.Printf("❌ Error creating dir %s/pipelines, %v", projectName, err)
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
				fmt.Printf("❌  Error reading template: %v", err)
				return
			}

			tmplName := filepath.Base(templateFile)
			tmpl, err := template.New(tmplName).Parse(string(tmplContent))

			if err != nil {
				fmt.Printf("❌  Error parsing template: %v", err)
				return
			}

			buffer := &bytes.Buffer{}

			err = tmpl.Execute(buffer, projectName)

			if err != nil {
				fmt.Printf("❌  Error executing template: '%s', %v", tmpl.Name(), err)
				return
			}

			formattedCode, err := format.Source(buffer.Bytes())

			if err != nil {
				fmt.Printf("❌  Error formatting sourcecode '%s', %v", tmpl.Name(), err)
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
				fmt.Printf("❌  Error creating %s.", sourceFilename)
				return
			}

			fmt.Printf("✔️  %s\n", sourceFilename)

		}
		fmt.Printf("\n✨ Congrates, %s created successfully.", projectName)
	},
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
		fmt.Printf("Directory '%s' already exists. Continue? (Y/N): ", dir)
		var input string
		_, err := fmt.Scan(&input)

		if err != nil {
			return err
		}

		if strings.ToLower(input) != "y" {
			return nil
		}
	}

	return os.MkdirAll(dir, os.ModePerm)
}
