package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/denwerk/moco/src/config"
)

func fetchProjects(cfg *config.Config) ([]Project, error) {
	url := fmt.Sprintf("https://%s.mocoapp.com/api/v1/projects/assigned", cfg.MocoDomain)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+cfg.MocoAPIKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, fmt.Errorf("error unmarshaling projects: %v", err)
	}

	return projects, nil
}

func fetchTimeEntries(cfg *config.Config, date string) ([]TimeEntry, error) {
	url := fmt.Sprintf("https://%s.mocoapp.com/api/v1/time_entries?from=%s&to=%s&user_id=current", cfg.MocoDomain, date, date)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+cfg.MocoAPIKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var entries []TimeEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("error unmarshaling time entries: %v", err)
	}

	return entries, nil
}

func submitTimeEntry(cfg *config.Config, entry TimeEntry) error {
	url := fmt.Sprintf("https://%s.mocoapp.com/api/v1/time_entries", cfg.MocoDomain)

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Token "+cfg.MocoAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to submit time entry: status code %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
