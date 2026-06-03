// neuralops is the official NeuralOps CLI v1 for governance and export automation.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const defaultBase = "http://localhost:8080/api/v1"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	base := envOr("NEURALOPS_API_URL", defaultBase)
	token := os.Getenv("NEURALOPS_API_TOKEN")

	switch os.Args[1] {
	case "version":
		fmt.Println("neuralops-cli v1.0.0")
	case "health":
		tenant := envOr("NEURALOPS_TENANT", "default")
		exitOn(runGET(strings.TrimSuffix(base, "/api/v1")+"/health", token, tenant))
	case "slos":
		tenant := envOr("NEURALOPS_TENANT", "default")
		exitOn(runGET(base+"/slos", token, tenant))
	case "gate":
		gate := "C"
		if len(os.Args) > 2 {
			gate = strings.ToUpper(os.Args[2])
		}
		fmt.Printf("Run: ./scripts/gate-verify.sh --gate %s\n", gate)
	case "msp-tenants":
		exitOn(runGET(base+"/admin/msp/tenants", token, ""))
	case "abac-policies":
		exitOn(runGET(base+"/admin/abac-policies", token, ""))
	case "export":
		fs := flag.NewFlagSet("export", flag.ExitOnError)
		kind := fs.String("type", "warehouse", "export type: warehouse|bi|events")
		dest := fs.String("destination", "", "destination URI (required)")
		_ = fs.Parse(os.Args[2:])
		if *dest == "" {
			fmt.Fprintln(os.Stderr, "destination is required")
			os.Exit(1)
		}
		path := base + "/exports/" + strings.ToLower(*kind)
		exitOn(runPOST(path, token, map[string]string{"destination": *dest}))
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`NeuralOps CLI v1

Usage:
  neuralops version
  neuralops health
  neuralops slos
  neuralops gate [B|C|D|E|F]
  neuralops msp-tenants
  neuralops abac-policies
  neuralops export -type warehouse|bi|events -destination <uri>

Environment:
  NEURALOPS_API_URL   API base (default http://localhost:8080/api/v1)
  NEURALOPS_API_TOKEN Optional bearer token
  NEURALOPS_TENANT    Tenant header for health/slos (default default)`)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return strings.TrimRight(v, "/")
	}
	return fallback
}

func exitOn(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runGET(url, token, tenant string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if tenant != "" {
		req.Header.Set("X-Tenant-ID", tenant)
	}
	setAuth(req, token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return printResponse(resp)
}

func runPOST(url, token string, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	setAuth(req, token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return printResponse(resp)
}

func setAuth(req *http.Request, token string) {
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func printResponse(resp *http.Response) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}
	var pretty map[string]any
	if json.Unmarshal(data, &pretty) == nil {
		out, _ := json.MarshalIndent(pretty, "", "  ")
		fmt.Println(string(out))
		return nil
	}
	fmt.Println(string(data))
	return nil
}
