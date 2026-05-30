package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "goblocks",
		Short: "Goblocks service framework CLI",
		Long:  "Scaffold Go services using the Goblocks framework with DDD onion architecture.",
	}

	root.AddCommand(newCmd())
	root.AddCommand(configCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
