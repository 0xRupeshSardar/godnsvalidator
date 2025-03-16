package config

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	Targets    []string
	Exclusions []string
	OutputFile string
	RootDomain string
	Query      string
	Threads    int
	Timeout    int
	NoColor    bool
	Verbose    bool
	Silent     bool
}

func ParseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.RootDomain, "r", "bet365.com", "Root domain for validation")
	flag.StringVar(&cfg.Query, "q", "dnsvalidator", "Validation query prefix")
	flag.IntVar(&cfg.Threads, "t", 5, "Number of worker threads")
	flag.IntVar(&cfg.Timeout, "timeout", 10, "DNS query timeout in seconds")
	flag.BoolVar(&cfg.NoColor, "no-color", false, "Disable color output")
	flag.BoolVar(&cfg.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&cfg.Silent, "silent", false, "Silent mode")
	flag.StringVar(&cfg.OutputFile, "o", "", "Output file")

	var targets, exclusions string
	flag.StringVar(&targets, "tL", "", "Target list (file or URL)")
	flag.StringVar(&exclusions, "eL", "", "Exclusion list (file or URL)")

	flag.Parse()

	cfg.Targets = loadTargets(targets)
	cfg.Exclusions = loadTargets(exclusions)

	return cfg
}

func loadTargets(source string) []string {
	if source == "" {
		return nil
	}

	if strings.HasPrefix(source, "http") {
		return loadHTTP(source)
	}
	return loadFile(source)
}

func loadFile(path string) []string {
    file, err := os.Open(path)
    if err != nil {
        fmt.Printf("Error opening file: %v\n", err)
        return nil
    }
    defer file.Close()

    var targets []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        if target := strings.TrimSpace(scanner.Text()); target != "" {
            targets = append(targets, target)
        }
    }
    if err := scanner.Err(); err != nil {
        fmt.Printf("Error reading file: %v\n", err)
    }
    return targets
}

func loadHTTP(url string) []string {
    resp, err := http.Get(url)
    if err != nil {
        fmt.Printf("Error fetching URL: %v\n", err)
        return nil
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        fmt.Printf("Error: got status code %d\n", resp.StatusCode)
        return nil
    }

    var targets []string
    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        if target := strings.TrimSpace(scanner.Text()); target != "" {
            targets = append(targets, target)
        }
    }
    if err := scanner.Err(); err != nil {
        fmt.Printf("Error reading response: %v\n", err)
    }
    return targets
}
