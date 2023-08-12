/*
Copyright © 2023 Tech Engine
*/
package cmd

import (
	"bytes"
	"fmt"
	"go/format"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/spf13/cobra"
)

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

		err = tmpl.Execute(buffer, pipelineName)

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

		sourceFilename := filepath.Join("pipelines", strings.TrimSuffix(tmpl.Name(), ".tmpl")+".go")

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

func capitalizeFirstLetter(s string) string {
	if s == "" {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])

	return string(r)
}
