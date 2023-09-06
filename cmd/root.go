/*
Copyright © 2023 Tech Engine
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const VERSION = "0.7.9"

const BANNER_MSG = `
_________       ________                                         
__  ____/______ __  ___/_____________________ _________ _____  __
_  / __  _  __ \_____ \ _  ___/__  ___/_  __ '/___  __ \__  / / /
/ /_/ /  / /_/ /____/ / / /__  _  /    / /_/ / __  /_/ /_  /_/ / 
\____/   \____/ /____/  \___/  /_/     \__,_/  _  .___/ _\__, /  
                                               /_/      /____/   

GoScrapy: Harnessing Go's power for efficient web scraping, inspired by Python's Scrapy framework.`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "goscrapy [command]",
	Short:   "A command line tool to everything related to GoScrapy.",
	Long:    BANNER_MSG,
	Version: VERSION,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Whoops :( !!! There was an error '%s'", err.Error())
		os.Exit(1)
	}
}

func init() {
}
