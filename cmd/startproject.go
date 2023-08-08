/*
Copyright © 2023 Tech Engine
*/
package cmd

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

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

		err = createDirIfNotExist(projectName)

		if err != nil {
			fmt.Printf("❌ Error creating dir '%s', %v", projectName, err)
			return
		}

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

			sourceFilename := filepath.Join(projectName, strings.TrimSuffix(tmpl.Name(), ".tmpl")+".go")

			err = writeToFile(sourceFilename, formattedCode)

			if err != nil {
				fmt.Printf("❌  Error creating %s.", sourceFilename)
				return
			}

			fmt.Printf("✔️  %s\n", sourceFilename)

		}
		fmt.Printf("\n✨ Congrates. %s created successfully.", projectName)
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
		fmt.Printf("Project directory '%s' already exists. Continue? (Y/N): ", dir)
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

func capitalizeFirstLetter(s string) string {
	if s == "" {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])

	return string(r)
}
