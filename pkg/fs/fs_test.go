package fs_test

import (
	"github.com/RealFax/RedQueen/pkg/fs"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestIsExist(t *testing.T) {
	existingPath := "./testdata/existing_file.txt"
	assert.True(t, fs.IsExist(existingPath), "Existing path should return true")

	nonExistingPath := "./testdata/non_existing_file.txt"
	assert.False(t, fs.IsExist(nonExistingPath), "Non-existing path should return false")
}

func TestMustOpenWithFlag(t *testing.T) {
	existingFilePath := "./testdata/existing_file.txt"
	file, err := fs.MustOpenWithFlag(existingFilePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE)
	assert.NoError(t, err)
	assert.NotNil(t, file)
	file.Close()

	nonExistingDirPath := "./testdata/non_existing_dir/file.txt"
	file, err = fs.MustOpenWithFlag(nonExistingDirPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE)
	assert.NoError(t, err)
	assert.NotNil(t, file)
	file.Close()
}

func TestMustOpen(t *testing.T) {
	existingFilePath := "./testdata/existing_file.txt"
	file, err := fs.MustOpen(existingFilePath)
	assert.NoError(t, err)
	assert.NotNil(t, file)
	file.Close()

	nonExistingDirPath := "./testdata/non_existing_dir/file.txt"
	file, err = fs.MustOpen(nonExistingDirPath)
	assert.NoError(t, err)
	assert.NotNil(t, file)
	file.Close()
}
