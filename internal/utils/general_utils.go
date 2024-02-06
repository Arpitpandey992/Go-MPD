package utils

import "os"

// checks if file exists in system, returns:
// bool: if it exists or not
// error: if any other error occurred
func DoesFileExistsInFileSystem(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		// File exists
		return true, nil
	} else if os.IsNotExist(err) {
		// File does not exist
		return false, nil
	}
	// An error occurred
	return false, err
}
