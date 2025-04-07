package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type LastTask struct {
	ProjectID int    `json:"project_id"`
	TaskID    int    `json:"task_id"`
	TaskTitle string `json:"task_title"`
}

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
	}

	configDir := filepath.Join(homeDir, ".moco")

	// Ensure directory exists with correct permissions
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("error creating config directory: %v", err)
	}

	// Ensure directory has correct permissions
	if err := os.Chmod(configDir, 0700); err != nil {
		return "", fmt.Errorf("error setting config directory permissions: %v", err)
	}

	return configDir, nil
}

func SaveLastTask(task LastTask) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "last_task.json")
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("error marshaling task data: %v", err)
	}

	// Write file with correct permissions
	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	// Ensure file has correct permissions
	if err := os.Chmod(configFile, 0600); err != nil {
		return fmt.Errorf("error setting config file permissions: %v", err)
	}

	return nil
}

func LoadLastTask() (*LastTask, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(configDir, "last_task.json")

	// Check if file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create empty file if it doesn't exist
		if err := os.WriteFile(configFile, []byte("{}"), 0600); err != nil {
			return nil, fmt.Errorf("error creating config file: %v", err)
		}
		return nil, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	// If file is empty, return nil
	if len(data) == 0 || string(data) == "{}" {
		return nil, nil
	}

	var task LastTask
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("error unmarshaling task data: %v", err)
	}

	return &task, nil
}
