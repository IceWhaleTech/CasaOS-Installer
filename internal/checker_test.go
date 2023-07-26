package internal_test

import (
	"os"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/stretchr/testify/assert"
)

func TestGetChecksum(t *testing.T) {
	// Create a temporary file.
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatal(err)
	}

	// Delete the temporary file.
	defer os.Remove(tmpfile.Name())

	// Write test data to the file.
	testData := `
# This is a comment.
1234567890abcdef filename1.txt
badc0ffee0ddf00d filename2.txt
	`
	if _, err := tmpfile.Write([]byte(testData)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Call the function with the path to the temporary file.
	checksums, err := internal.GetChecksums(tmpfile.Name())

	// Assert that the function did not return an error.
	assert.NoError(t, err)

	// Assert that the function returned the expected checksums.
	expectedChecksums := map[string]string{
		"filename1.txt": "1234567890abcdef",
		"filename2.txt": "badc0ffee0ddf00d",
	}
	assert.Equal(t, expectedChecksums, checksums)
}
