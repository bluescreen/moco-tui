package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/denwerk/moco/src/config"
	"github.com/denwerk/moco/src/logger"
	"github.com/denwerk/moco/src/types"
)

func fetchProjects(cfg *config.Config) ([]types.Project, error) {
	url := fmt.Sprintf("https://%s.mocoapp.com/api/v1/projects/assigned", cfg.MocoDomain)
	logger.LogAPIRequest("GET", url, nil)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.LogAPIError(err)
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+cfg.MocoAPIKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.LogAPIError(err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.LogAPIError(err)
		return nil, err
	}

	logger.LogAPIResponse(resp.StatusCode, body)

	var projects []types.Project
	if err := json.Unmarshal(body, &projects); err != nil {
		logger.LogAPIError(err)
		return nil, fmt.Errorf("error unmarshaling projects: %v", err)
	}

	return projects, nil
}

func fetchTimeEntries(cfg *config.Config, date string) ([]types.TimeEntry, error) {
	// Parse the input date
	endDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		logger.LogAPIError(err)
		return nil, err
	}

	// Calculate the start date (6 days before end date)
	startDate := endDate.AddDate(0, 0, -6)

	// Format dates for API
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	url := fmt.Sprintf("https://%s.mocoapp.com/api/v1/activities?from=%s&to=%s", cfg.MocoDomain, startDateStr, endDateStr)
	logger.LogAPIRequest("GET", url, nil)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.LogAPIError(err)
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+cfg.MocoAPIKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.LogAPIError(err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.LogAPIError(err)
		return nil, err
	}

	logger.LogAPIResponse(resp.StatusCode, body)

	var entries []types.TimeEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		logger.LogAPIError(err)
		return nil, fmt.Errorf("error unmarshaling time entries: %v", err)
	}

	return entries, nil
}

func submitTimeEntry(cfg *config.Config, entry types.TimeEntry) error {
	url := fmt.Sprintf("https://%s.mocoapp.com/api/v1/activities", cfg.MocoDomain)

	jsonData, err := json.Marshal(entry)
	if err != nil {
		logger.LogAPIError(err)
		return err
	}

	logger.LogAPIRequest("POST", url, jsonData)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		logger.LogAPIError(err)
		return err
	}

	req.Header.Set("Authorization", "Token "+cfg.MocoAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.LogAPIError(err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.LogAPIError(err)
		return err
	}

	logger.LogAPIResponse(resp.StatusCode, body)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to submit time entry: status code %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func deleteTimeEntry(cfg *config.Config, id int) error {
	url := fmt.Sprintf("https://%s.mocoapp.com/api/v1/activities/%d", cfg.MocoDomain, id)
	logger.LogAPIRequest("DELETE", url, nil)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		logger.LogAPIError(err)
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Token "+cfg.MocoAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.LogAPIError(err)
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.LogAPIError(fmt.Errorf("error deleting time entry: %s", string(body)))
		return fmt.Errorf("error deleting time entry: %s", string(body))
	}

	return nil
}
