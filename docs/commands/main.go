package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra/doc"
	"github.com/yaroslav-koval/hange/cmd"
)

func main() {
	rootCmd := cmd.GetRootCmd()

	dir := flag.String("docs-dir", ".", "directory for generated commands documentation")
	flag.Parse()

	fmt.Println("Output directory:", *dir)

	err := doc.GenMarkdownTree(rootCmd, *dir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
