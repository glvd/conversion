package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

const programName = `conversion`

var rootCmd = &cobra.Command{
	Use:     "conversion [command]",
	Short:   "call conversion command",
	Version: "v0.0.1",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("arguments [command] was not inputted")
	},
	DisableSuggestions:         false,
	SuggestionsMinimumDistance: 1,
}

func main() {
	rootCmd.AddCommand()
	e := rootCmd.Execute()
	if e != nil {
		panic(e)
	}
}
