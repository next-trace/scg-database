// Package main provides the SCG database CLI tool with commands for generating
// database models, migrations, and other database-related files.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	makeCmd = &cobra.Command{
		Use:   "make",
		Short: "Commands to make new files",
	}

	makeModelCmd = &cobra.Command{
		Use:   "model [ModelName]",
		Short: "Create a new model file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd
			modelName := args[0]
			modelsPath := viper.GetString("database.paths.models")
			if modelsPath == "" {
				fmt.Println("Error: 'database.paths.models' not set in config file.")
				os.Exit(1)
			}

			dirPath := filepath.Join(modelsPath, strings.ToLower(modelName))
			filePath := filepath.Join(dirPath, strings.ToLower(modelName)+".go")

			// Validate directory path to prevent directory traversal
			if err := validatePath(modelsPath, dirPath); err != nil {
				fmt.Printf("Error: invalid model directory path: %v\n", err)
				os.Exit(1)
			}

			if _, err := os.Stat(filePath); !os.IsNotExist(err) {
				fmt.Printf("Error: Model file already exists at %s\n", filePath)
				os.Exit(1)
			}

			if err := os.MkdirAll(dirPath, 0o750); err != nil {
				fmt.Printf("Error creating model directory: %v\n", err)
				os.Exit(1)
			}

			tmpl, err := template.New("model").Parse(modelTemplate)
			if err != nil {
				panic(err) // Should not happen
			}

			file, err := secureCreateFile(modelsPath, filePath)
			if err != nil {
				fmt.Printf("Error creating model file: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()

			caser := cases.Title(language.English)
			data := struct {
				ModelName   string
				PackageName string
			}{
				ModelName:   caser.String(modelName),
				PackageName: strings.ToLower(modelName),
			}

			if err := tmpl.Execute(file, data); err != nil {
				fmt.Printf("Error executing template: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Model created: %s\n", filePath)
		},
	}
)

func init() {
	makeCmd.AddCommand(makeModelCmd)
}

const (
	modelTemplate = `package {{.PackageName}}

import (
	"github.com/next-trace/scg-database/contract"
)

type {{.ModelName}} struct {
	contract.BaseModel
	// Define your fields here
}

// TableName returns the database table name for the {{.ModelName}} model.
func (m *{{.ModelName}}) TableName() string {
	// Return the table name for this model
	// Example: return "{{.PackageName}}s"
	return ""
}

// Relationships defines the relationships for the {{.ModelName}} model.
func (m *{{.ModelName}}) Relationships() map[string]contract.Relationship {
	return map[string]contract.Relationship{
		// Example relationships:
		// "Profile": contract.NewHasOne(&Profile{}, "user_id", "id"),
		// "Orders": contract.NewHasMany(&Order{}, "user_id", "id"),
		// "Roles": contract.NewBelongsToMany(&Role{}, "user_roles"),
	}
}
`
)
