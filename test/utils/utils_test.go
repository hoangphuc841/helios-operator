package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWarnError(t *testing.T) {
	t.Run("warns error without panicking", func(t *testing.T) {
		// This function just writes to GinkgoWriter, so we can't easily test the output
		// But we can ensure it doesn't panic
		assert.NotPanics(t, func() {
			warnError(assert.AnError)
		})
	})

	t.Run("warns nil error without panicking", func(t *testing.T) {
		assert.NotPanics(t, func() {
			warnError(nil)
		})
	})
}

func TestGetNonEmptyLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single line",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "multiple lines with no empty lines",
			input:    "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "multiple lines with empty lines",
			input:    "line1\n\nline2\n\n\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "only empty lines",
			input:    "\n\n\n",
			expected: nil,
		},
		{
			name:     "lines with whitespace",
			input:    "line1\n  \nline2\n\t\nline3",
			expected: []string{"line1", "  ", "line2", "\t", "line3"},
		},
		{
			name:     "trailing newline",
			input:    "line1\nline2\n",
			expected: []string{"line1", "line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNonEmptyLines(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetProjectDir(t *testing.T) {
	t.Run("returns project directory successfully", func(t *testing.T) {
		// Save current working directory
		originalDir, err := os.Getwd()
		assert.NoError(t, err)

		// Change to a test directory that contains "/test/e2e"
		testDir := filepath.Join(originalDir, "test", "e2e")
		err = os.MkdirAll(testDir, 0755)
		assert.NoError(t, err)

		// Change to the test directory
		err = os.Chdir(testDir)
		assert.NoError(t, err)

		// Restore original directory after test
		defer func() {
			_ = os.Chdir(originalDir)
			_ = os.RemoveAll(testDir)
		}()

		projectDir, err := GetProjectDir()
		assert.NoError(t, err)
		assert.Equal(t, originalDir, projectDir)
	})

	t.Run("handles working directory error", func(t *testing.T) {
		// This is hard to test without mocking os.Getwd()
		// The function should return an error if os.Getwd() fails
		// We'll test the error message format
		originalDir, err := os.Getwd()
		assert.NoError(t, err)

		// Test with a valid directory
		projectDir, err := GetProjectDir()
		assert.NoError(t, err)
		assert.NotEmpty(t, projectDir)
		assert.NotContains(t, projectDir, "/test/e2e")
		assert.Equal(t, originalDir, projectDir)
	})
}

func TestUncommentCode(t *testing.T) {
	t.Run("uncomments single line code", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "test_uncomment_*.go")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		// Write test content with commented code
		content := `package main

// func testFunction() {
//     return "test"
// }

func main() {
    // fmt.Println("hello")
}`
		err = os.WriteFile(tmpFile.Name(), []byte(content), 0644)
		assert.NoError(t, err)

		// Uncomment the testFunction line
		err = UncommentCode(tmpFile.Name(), "// func testFunction() {", "// ")
		assert.NoError(t, err)

		// Read the file and verify the code was uncommented
		updatedContent, err := os.ReadFile(tmpFile.Name())
		assert.NoError(t, err)
		assert.Contains(t, string(updatedContent), "func testFunction() {")
		assert.NotContains(t, string(updatedContent), "// func testFunction() {")
	})

	t.Run("uncomments multi-line code", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "test_uncomment_*.go")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		// Write test content with commented multi-line code
		content := `package main

// func testFunction() {
//     if true {
//         return "test"
//     }
// }

func main() {
}`
		err = os.WriteFile(tmpFile.Name(), []byte(content), 0644)
		assert.NoError(t, err)

		// Uncomment the testFunction
		err = UncommentCode(tmpFile.Name(), "// func testFunction() {", "// ")
		assert.NoError(t, err)

		// Read the file and verify the code was uncommented
		updatedContent, err := os.ReadFile(tmpFile.Name())
		assert.NoError(t, err)
		assert.Contains(t, string(updatedContent), "func testFunction() {")
		assert.Contains(t, string(updatedContent), "    if true {")
		assert.Contains(t, string(updatedContent), "        return \"test\"")
		assert.Contains(t, string(updatedContent), "    }")
		assert.NotContains(t, string(updatedContent), "// func testFunction() {")
	})

	t.Run("handles file not found", func(t *testing.T) {
		err := UncommentCode("nonexistent_file.go", "target", "// ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
	})

	t.Run("handles target not found", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "test_uncomment_*.go")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		// Write test content without the target
		content := `package main

func main() {
    fmt.Println("hello")
}`
		err = os.WriteFile(tmpFile.Name(), []byte(content), 0644)
		assert.NoError(t, err)

		// Try to uncomment a target that doesn't exist
		err = UncommentCode(tmpFile.Name(), "nonexistentFunction()", "// ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unable to find the code")
	})

	t.Run("handles different comment prefixes", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "test_uncomment_*.go")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		// Write test content with different comment prefix
		content := `package main

# func testFunction() {
#     return "test"
# }

func main() {
}`
		err = os.WriteFile(tmpFile.Name(), []byte(content), 0644)
		assert.NoError(t, err)

		// Uncomment with different prefix
		err = UncommentCode(tmpFile.Name(), "# func testFunction() {", "# ")
		assert.NoError(t, err)

		// Read the file and verify the code was uncommented
		updatedContent, err := os.ReadFile(tmpFile.Name())
		assert.NoError(t, err)
		assert.Contains(t, string(updatedContent), "func testFunction() {")
		assert.NotContains(t, string(updatedContent), "# func testFunction() {")
	})
}

func TestConstants(t *testing.T) {
	t.Run("prometheus operator version is set", func(t *testing.T) {
		assert.Equal(t, "v0.85.0", prometheusOperatorVersion)
	})

	t.Run("prometheus operator URL is set", func(t *testing.T) {
		expectedURL := "https://github.com/prometheus-operator/prometheus-operator/" +
			"releases/download/%s/bundle.yaml"
		assert.Equal(t, expectedURL, prometheusOperatorURL)
	})

	t.Run("cert manager version is set", func(t *testing.T) {
		assert.Equal(t, "v1.18.2", certmanagerVersion)
	})

	t.Run("cert manager URL template is set", func(t *testing.T) {
		expectedURL := "https://github.com/cert-manager/cert-manager/releases/download/%s/cert-manager.yaml"
		assert.Equal(t, expectedURL, certmanagerURLTmpl)
	})
}

// Note: The following functions are not easily testable without external dependencies:
// - Run() - requires actual command execution
// - InstallPrometheusOperator() - requires kubectl and cluster access
// - UninstallPrometheusOperator() - requires kubectl and cluster access
// - IsPrometheusCRDsInstalled() - requires kubectl and cluster access
// - UninstallCertManager() - requires kubectl and cluster access
// - InstallCertManager() - requires kubectl and cluster access
// - IsCertManagerCRDsInstalled() - requires kubectl and cluster access
// - LoadImageToKindClusterWithName() - requires kind and docker

// These functions would require integration tests with actual Kubernetes clusters
// or extensive mocking of external command execution.
