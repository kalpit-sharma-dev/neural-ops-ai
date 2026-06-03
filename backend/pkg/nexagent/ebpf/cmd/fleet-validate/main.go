// Fleet validation CLI — exits 0 when the host passes the eBPF rollout matrix.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/neuralops/platform/pkg/nexagent/ebpf/validate"
)

func main() {
	report := validate.ProbeHost()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "encode report: %v\n", err)
		os.Exit(2)
	}
	if !report.Passed {
		os.Exit(1)
	}
}
