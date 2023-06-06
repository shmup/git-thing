package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <repository_path>")
		return
	}
	repoPath := os.Args[1]

	numCommits := 10000

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < numCommits; i++ {
		yamlFile, err := getRandomYAMLFile(repoPath)
		if err != nil {
			log.Fatalf("Failed to get random YAML file: %v", err)
		}

		if err := changeRandomLine(yamlFile); err != nil {
			log.Fatalf("Failed to change line in YAML file: %v", err)
		}

		if err := commitChanges(repoPath, yamlFile); err != nil {
			log.Fatalf("Failed to commit changes: %v", err)
		}

		fmt.Printf("Commit %d created\n", i+1)
	}
}

func getRandomYAMLFile(dirPath string) (string, error) {
	var yamlFiles []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
			yamlFiles = append(yamlFiles, path)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if len(yamlFiles) == 0 {
		return "", fmt.Errorf("no YAML files found in %s", dirPath)
	}

	randomIndex := rand.Intn(len(yamlFiles))
	return yamlFiles[randomIndex], nil
}

func changeRandomLine(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	lines, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	allLines := strings.Split(string(lines), "\n")
	var candidateLines []int

	for i, line := range allLines {
		if strings.HasPrefix(strings.TrimSpace(line), "imageName") {
			candidateLines = append(candidateLines, i)
		}
	}

	if len(candidateLines) == 0 {
		return fmt.Errorf("no lines starting with 'imageName' found in %s", filePath)
	}

	randomIndex := rand.Intn(len(candidateLines))
	lineIndex := candidateLines[randomIndex]
	line := allLines[lineIndex]

	randNumber := fmt.Sprintf("%04d", rand.Intn(10000))

	if len(line) > 4 {
		line = line[:len(line)-4] + randNumber
	} else {
		line = randNumber
	}

	allLines[lineIndex] = line

	updatedContent := strings.Join(allLines, "\n")
	if err := file.Truncate(0); err != nil {
		return err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	if _, err := file.WriteString(updatedContent); err != nil {
		return err
	}

	return nil
}

func commitChanges(repoPath, filePath string) error {
	cmd := exec.Command("git", "commit", "-am", fmt.Sprintf("Updated %s", filePath))
	cmd.Dir = repoPath

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
