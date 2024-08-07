package utils

import (
	"io"
	"os"
)

// ReadEntireFile reads the entire contents of a file into memory
func ReadEntireFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()

	// Read the contents of the file
	testData, err := io.ReadAll(f)
	if err != nil {
		return []byte{}, err
	}
	return testData, nil
}

// WriteEntireFile writes the entire contents of a file to disk
func WriteEntireFile(path string, data []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}
