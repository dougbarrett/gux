package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	githubRepo   = "dougbarrett/gux"
	githubAPIURL = "https://api.github.com/repos/" + githubRepo + "/releases/latest"
	modulePath   = "github.com/" + githubRepo + "/cmd/gux"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// checkForUpdates checks for updates and prints a warning if outdated.
// Uses a short timeout to avoid slowing down commands.
func checkForUpdates() {
	currentVersion := getVersion()

	// Skip check for dev versions to avoid noise during development
	if currentVersion == "dev" || currentVersion == "(devel)" {
		return
	}

	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return // Silently fail
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "gux-cli")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return // Silently fail
	}
	defer resp.Body.Close()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	if !isUpToDate(currentVersion, release.TagName) {
		fmt.Printf("\nUpdate available: %s -> %s (run 'gux update' to upgrade)\n", currentVersion, release.TagName)
	}
}

func runUpdate(checkOnly bool) {
	currentVersion := getVersion()
	fmt.Printf("Current version: %s\n", currentVersion)

	// Fetch latest release from GitHub
	fmt.Println("Checking for updates...")

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "gux-cli")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching release info: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		fmt.Println("No releases found. You may be running a development version.")
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("GitHub API returned status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Printf("Error parsing release info: %v\n", err)
		os.Exit(1)
	}

	latestVersion := release.TagName
	fmt.Printf("Latest version:  %s\n", latestVersion)

	// Compare versions
	if isUpToDate(currentVersion, latestVersion) {
		fmt.Println("\nYou're already on the latest version!")
		return
	}

	fmt.Printf("\nNew version available: %s\n", latestVersion)
	fmt.Printf("Release: %s\n", release.HTMLURL)

	if checkOnly {
		fmt.Printf("\nRun 'gux update' to install the latest version.\n")
		return
	}

	// Perform update using go install
	fmt.Printf("\nUpdating to %s...\n", latestVersion)

	cmd := exec.Command("go", "install", modulePath+"@"+latestVersion)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError updating: %v\n", err)
		fmt.Println("\nYou can manually update by running:")
		fmt.Printf("  go install %s@%s\n", modulePath, latestVersion)
		os.Exit(1)
	}

	fmt.Printf("\nSuccessfully updated to %s!\n", latestVersion)
}

// isUpToDate compares current and latest versions
func isUpToDate(current, latest string) bool {
	// Normalize versions (remove 'v' prefix if present)
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// Handle dev version
	if current == "dev" || current == "(devel)" {
		return false // Always offer update for dev versions
	}

	return current == latest
}
