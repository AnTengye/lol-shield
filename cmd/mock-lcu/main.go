package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/AnTengye/lol-shield/internal/mocklcu"
)

type config struct {
	addr        string
	scenarioDir string
}

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

	scenario, err := mocklcu.LoadScenario(cfg.scenarioDir)
	if err != nil {
		return fmt.Errorf("load scenario: %w", err)
	}

	return http.ListenAndServe(cfg.addr, mocklcu.NewServer(scenario))
}

func parseConfig(args []string) (config, error) {
	fs := flag.NewFlagSet("mock-lcu", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	cfg := config{}
	fs.StringVar(&cfg.addr, "addr", "127.0.0.1:19365", "listen address")
	fs.StringVar(&cfg.scenarioDir, "scenario-dir", "internal/mocklcu/fixtures/default", "scenario fixture directory")

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}

	return cfg, nil
}
