/*
Copyright © 2023 Tech Engine
*/
package cli

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/spf13/cobra"
)

const maxSearchDepth = 4

var pipelineCmd = &cobra.Command{
	Use:   "pipeline [pipeline-name]",
	Short: "Creates a new GoScrapy pipeline with the specified name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var pipelineName = strings.TrimSpace(args[0])

		if pipelineName == "" {
			fmt.Printf("⚠️ please provide a pipeline-name")
			return
		}

		// read template
		tmplContent, err := templatesFS.ReadFile("templates/pipeline.tmpl")

		if err != nil {
			fmt.Printf("❌  Error reading template: %v", err)
			return
		}

		tmpl, err := template.New(pipelineName).
			Funcs(template.FuncMap{
				"capitalizeFirstLetter": capitalizeFirstLetter,
			}).Parse(string(tmplContent))

		if err != nil {
			fmt.Printf("❌  Error parsing template: %v", err)
			return
		}

		buffer := &bytes.Buffer{}

		err = tmpl.Execute(buffer, removeSpecialChars(pipelineName))

		if err != nil {
			fmt.Printf("❌  Error executing template: '%s', %v", tmpl.Name(), err)
			return
		}

		// formate golang code
		formattedCode, err := format.Source(buffer.Bytes())

		if err != nil {
			fmt.Printf("❌  Error formatting sourcecode '%s', %v", tmpl.Name(), err)
			return
		}

		// determine the correct path for pipelines
		targetDir, found := findPipelinesPath()
		if !found {
			fmt.Println("❌ Error: a project is required to use the pipeline command")
			return
		}

		sourceFilename := filepath.Join(targetDir, strings.TrimSuffix(tmpl.Name(), ".tmpl")+".go")

		// ensure pipelines dir exists
		if err := createDirIfNotExist(targetDir); err != nil {
			fmt.Printf("❌  Error creating pipelines directory: %v", err)
			return
		}

		// write go file
		err = writeToFile(sourceFilename, formattedCode)

		if err != nil {
			fmt.Printf("❌  Error creating %s.", sourceFilename)
			return
		}

		fmt.Printf("✔️  %s\n", sourceFilename)

		fmt.Printf("\n✨ Congrates, %s created successfully.", pipelineName)
	},
}

func init() {
	rootCmd.AddCommand(pipelineCmd)
}

// findPipelinesPath attempts to locate the correct pipelines directory within a GoScrapy project.
// Returns the path and a boolean indicating if a valid project structure was found.
func findPipelinesPath() (string, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", false
	}

	// Search upwards for go.mod to find the project root
	root := cwd
	hasMod := false
	for depth := 0; depth < maxSearchDepth; depth++ {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			hasMod = true
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			break
		}
		root = parent
	}

	if !hasMod {
		return "", false
	}

	// Search for the project source subdirectory inside the root
	// We look for a folder that contains 'base.go' and an existing 'pipelines' folder
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subPath := filepath.Join(root, entry.Name())
			// Check if this subfolder is the project folder
			if _, err := os.Stat(filepath.Join(subPath, "base.go")); err == nil {
				// We found a folder with base.go, return its pipelines subfolder
				return filepath.Join(subPath, "pipelines"), true
			}
		}
	}

	// Fallback: check if we are already inside a pipelines directory AND it's part of a project
	// Check if parent has base.go
	parent := filepath.Dir(cwd)
	if filepath.Base(cwd) == "pipelines" {
		if _, err := os.Stat(filepath.Join(parent, "base.go")); err == nil {
			return ".", true
		}
	}

	return "", false
}

func capitalizeFirstLetter(s string) string {
	if s == "" {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])

	return string(r)
}

func removeSpecialChars(input string) string {
	// Define a regular expression to match non-alphanumeric characters
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")

	// Replace matched characters with an empty string
	result := reg.ReplaceAllString(input, "")

	return result
}
