package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/AnTengye/lol-shield/internal/mocklcu"
)

type config struct {
	addr        string
	scenarioDir string
}

var (
	loadScenario   = mocklcu.LoadScenario
	newServer      = mocklcu.NewServer
	listenAndServe = http.ListenAndServe
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	cfg, err := parseConfig(args)
	if err != nil {
		return err
	}

	scenario, err := loadScenario(cfg.scenarioDir)
	if err != nil {
		return fmt.Errorf("load scenario: %w", err)
	}

	return listenAndServe(cfg.addr, newServer(scenario))
}

func parseConfig(args []string) (config, error) {
	fs := flag.NewFlagSet("mock-lcu", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	cfg := config{}
	fs.StringVar(&cfg.addr, "addr", "127.0.0.1:19365", "listen address")
	fs.StringVar(&cfg.scenarioDir, "scenario-dir", defaultScenarioDir(), "scenario fixture directory")

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}

	return cfg, nil
}

func defaultScenarioDir() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("internal", "mocklcu", "fixtures", "default")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "internal", "mocklcu", "fixtures", "default"))
}
