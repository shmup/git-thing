package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <repository_path> [num_commits]")
		return
	}
	repoPath := os.Args[1]

	numCommits := 100
	if len(os.Args) >= 3 {
		if n, err := strconv.Atoi(os.Args[2]); err == nil {
			numCommits = n
		}
	}

	rand.Seed(time.Now().UnixNano())

	startSize := getGitSize(repoPath)
	var endSize int64
	var totalIO time.Duration
	totalStart := time.Now()

	fmt.Printf("%-7s  %6s  %12s  %8s  %s\n", "sha", "commit", "size", "io", "delta")
	var prevIO time.Duration
	for i := 0; i < numCommits; i++ {
		yamlFile, err := getRandomYAMLFile(repoPath)
		if err != nil {
			log.Fatalf("Failed to get random YAML file: %v", err)
		}

		if err := changeRandomLine(yamlFile); err != nil {
			log.Fatalf("Failed to change line in YAML file: %v", err)
		}

		ioTime, err := commitChanges(repoPath, yamlFile)
		if err != nil {
			log.Fatalf("Failed to commit changes: %v", err)
		}
		totalIO += ioTime

		sha := getHeadSha(repoPath)
		endSize = getGitSize(repoPath)

		delta := ""
		if i > 0 {
			diff := ioTime - prevIO
			if diff > 0 {
				delta = fmt.Sprintf("\033[31m+%s\033[0m", fmtDuration(diff))
			} else if diff < 0 {
				delta = fmt.Sprintf("\033[32m%s\033[0m", fmtDuration(diff))
			} else {
				delta = "="
			}
		}
		prevIO = ioTime

		fmt.Printf("%-7s  %6d  %12s  %8s  %s\n", sha, i+1, humanSize(endSize), fmtDuration(ioTime), delta)
	}

	totalTime := time.Since(totalStart)
	overhead := totalTime - totalIO
	ioPct := float64(totalIO) / float64(totalTime) * 100

	growthPerCommit := float64(endSize-startSize) / float64(numCommits)
	estAt1M := startSize + int64(growthPerCommit*1_000_000)
	fmt.Printf("\ngrowth: ~%s/commit, est @ 1M: %s\n", humanSize(int64(growthPerCommit)), humanSize(estAt1M))
	fmt.Printf("time: %s total, %s io (%.0f%%), %s overhead\n", fmtDuration(totalTime), fmtDuration(totalIO), ioPct, fmtDuration(overhead))
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

func commitChanges(repoPath, filePath string) (time.Duration, error) {
	cmd := exec.Command("git", "commit", "-am", fmt.Sprintf("Updated %s", filePath))
	cmd.Dir = repoPath

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	if err != nil {
		return elapsed, err
	}

	return elapsed, nil
}

func getHeadSha(repoPath string) string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "???????"
	}
	return strings.TrimSpace(string(out))
}

func getGitSize(repoPath string) int64 {
	var size int64
	gitPath := filepath.Join(repoPath, ".git")
	filepath.Walk(gitPath, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func humanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.3f%c", float64(b)/float64(div), "KMGTPE"[exp])
}

func fmtDuration(d time.Duration) string {
	neg := d < 0
	if neg {
		d = -d
	}
	prefix := ""
	if neg {
		prefix = "-"
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%s%dµs", prefix, d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%s%dms", prefix, d.Milliseconds())
	}
	return fmt.Sprintf("%s%.1fs", prefix, d.Seconds())
}
