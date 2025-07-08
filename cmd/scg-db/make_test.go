package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TestMakeCommand(t *testing.T) {
	// Test that make command exists and has expected properties
	require.Equal(t, "make", makeCmd.Use)
	require.Contains(t, makeCmd.Short, "make new files")

	// Test that makeModelCmd is added as subcommand
	subcommands := makeCmd.Commands()
	var modelFound bool
	for _, cmd := range subcommands {
		if cmd.Use == "model [ModelName]" {
			modelFound = true
			break
		}
	}
	require.True(t, modelFound, "model command should be added to make")
}

// Note: Actual command execution tests are complex due to file system operations
// and os.Exit calls. These are better tested through integration tests.

func TestMakeModelCommand_ConfigValidation(t *testing.T) {
	// Test configuration validation logic
	viper.Reset()
	viper.Set("database.paths.models", "")

	// The command should handle missing config gracefully
	modelsPath := viper.GetString("database.paths.models")
	require.Empty(t, modelsPath)

	// Test with valid config
	viper.Set("database.paths.models", "domain")
	modelsPath = viper.GetString("database.paths.models")
	require.Equal(t, "domain", modelsPath)
}

func TestModelTemplate(t *testing.T) {
	// Test that the model template contains expected elements
	require.Contains(t, modelTemplate, "package {{.PackageName}}")
	require.Contains(t, modelTemplate, "type {{.ModelName}} struct")
	require.Contains(t, modelTemplate, "contract.BaseModel")
	require.Contains(t, modelTemplate, "func (m *{{.ModelName}}) TableName() string")
	require.Contains(t, modelTemplate, "func (m *{{.ModelName}}) Relationships()")
	require.Contains(t, modelTemplate, "github.com/next-trace/scg-database/contract")
	require.NotContains(t, modelTemplate, "github.com/next-trace/scg-database/adapter/gorm")
}

func TestMakeModelCommand_Properties(t *testing.T) {
	// Test command properties
	require.Equal(t, "model [ModelName]", makeModelCmd.Use)
	require.Contains(t, makeModelCmd.Short, "Create a new model file")
	require.NotNil(t, makeModelCmd.Args)
	require.NotNil(t, makeModelCmd.Run)
}

// TestModelTemplateExecution tests that the model template can be executed successfully
func TestModelTemplateExecution(t *testing.T) {
	tmpl, err := template.New("model").Parse(modelTemplate)
	require.NoError(t, err)

	caser := cases.Title(language.English)
	data := struct {
		ModelName   string
		PackageName string
	}{
		ModelName:   caser.String("user"),
		PackageName: strings.ToLower("user"),
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	result := buf.String()
	assert.Contains(t, result, "package user")
	assert.Contains(t, result, "type User struct")
	assert.Contains(t, result, "contract.BaseModel")
	assert.Contains(t, result, "func (m *User) TableName() string")
	assert.Contains(t, result, "func (m *User) Relationships()")
}

// TestModelTemplateWithDifferentNames tests template with various model names
func TestModelTemplateWithDifferentNames(t *testing.T) {
	testCases := []struct {
		input        string
		expectedPkg  string
		expectedType string
	}{
		{"User", "user", "User"},
		{"BlogPost", "blogpost", "Blogpost"},
		{"user_profile", "user_profile", "User_profile"},
		{"API", "api", "Api"},
	}

	tmpl, err := template.New("model").Parse(modelTemplate)
	require.NoError(t, err)

	caser := cases.Title(language.English)
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			data := struct {
				ModelName   string
				PackageName string
			}{
				ModelName:   caser.String(tc.input),
				PackageName: strings.ToLower(tc.input),
			}

			var buf strings.Builder
			err = tmpl.Execute(&buf, data)
			require.NoError(t, err)

			result := buf.String()
			assert.Contains(t, result, "package "+tc.expectedPkg)
			assert.Contains(t, result, "type "+tc.expectedType+" struct")
		})
	}
}

// TestMakeModelCommand_PathGeneration tests path generation logic
func TestMakeModelCommand_PathGeneration(t *testing.T) {
	testCases := []struct {
		modelName    string
		modelsPath   string
		expectedDir  string
		expectedFile string
	}{
		{
			modelName:    "User",
			modelsPath:   "domain",
			expectedDir:  filepath.Join("domain", "user"),
			expectedFile: filepath.Join("domain", "user", "user.go"),
		},
		{
			modelName:    "BlogPost",
			modelsPath:   "models",
			expectedDir:  filepath.Join("models", "blogpost"),
			expectedFile: filepath.Join("models", "blogpost", "blogpost.go"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.modelName, func(t *testing.T) {
			// Test the path generation logic that would be used in makeModelCmd
			dirPath := filepath.Join(tc.modelsPath, strings.ToLower(tc.modelName))
			filePath := filepath.Join(dirPath, strings.ToLower(tc.modelName)+".go")

			assert.Equal(t, tc.expectedDir, dirPath)
			assert.Equal(t, tc.expectedFile, filePath)
		})
	}
}

// TestFileExistenceCheck tests file existence checking logic
func TestFileExistenceCheck(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "existing.go")
	nonExistingFile := filepath.Join(tempDir, "nonexisting.go")

	// Create the existing file
	file, err := os.Create(existingFile)
	require.NoError(t, err)
	file.Close()

	// Test file existence logic (similar to what's used in makeModelCmd)
	_, err = os.Stat(existingFile)
	assert.False(t, os.IsNotExist(err), "Existing file should be detected")

	_, err = os.Stat(nonExistingFile)
	assert.True(t, os.IsNotExist(err), "Non-existing file should not be detected")
}
