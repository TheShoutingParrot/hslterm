package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	noApiKeyGiven = errors.New("no api key given (please set it, for help use -h)")
)

func apikeyFilePath() string {
	var path string
	if runtime.GOOS == "windows" {
		// On Windows, use %APPDATA%\hslterm
		roaming := os.Getenv("APPDATA")
		path = filepath.Join(roaming, appName, "apikey.txt")
	} else {
		// On Linux/macOS, use ~/.config/hslterm
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error finding home directory:", err)
			return ""
		}
		path = filepath.Join(home, ".config", appName, "apikey.txt")
	}
	return path
}

func saveApiKey(key string) error {
	filePath := apikeyFilePath()
	dir := filepath.Dir(filePath)
	// Ensure the directory exists
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(key)
	return err
}

func loadApikey() (string, error) {
	filePath := apikeyFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", noApiKeyGiven
		}
		return "", err
	}

	return string(data), err
}
