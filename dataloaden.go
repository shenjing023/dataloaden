package main

import (
	"fmt"
	"os"

	"github.com/shenjing023/dataloaden/pkg/generator"
	"github.com/spf13/cobra"
)

func main() {
	var (
		name      string
		keyType   string
		valueType string
		pkg       string
		dir       string
		rootCmd   = &cobra.Command{
			Use:   "dataloaden",
			Short: "Generate dataloader template file",
		}
		genCmd = &cobra.Command{
			Use:   "generate",
			Short: "Generate dataloader template file",
			Run: func(cmd *cobra.Command, args []string) {
				if dir == "" {
					wd, err := os.Getwd()
					if err != nil {
						fmt.Fprintln(os.Stderr, err.Error())
						os.Exit(2)
					}
					dir = wd
				}
				if err := generator.Generate(name, keyType, valueType, pkg, dir); err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					os.Exit(2)
				}
			},
		}
	)
	genCmd.Flags().StringVarP(&name, "struct", "s", "DataLoader", "set dataloader struct name")
	genCmd.Flags().StringVarP(&keyType, "keyType", "k", "int", "set function param value type eg. int int64 string")
	genCmd.Flags().StringVarP(&valueType, "valueType", "v", "int", "set function return value type eg. int int64 string")
	genCmd.Flags().StringVarP(&pkg, "package", "p", "dataloader", "set package name")
	genCmd.Flags().StringVarP(&dir, "directory", "d", "", "set generate dir path")

	rootCmd.AddCommand(genCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
