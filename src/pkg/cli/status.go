package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// ShowStatus queries the local health endpoint and prints a summary.
// It returns an error when the server is unreachable or unhealthy so the
// caller can exit non-zero, which is what container health checks rely on.
func ShowStatus(address string, port int) error {
	url := fmt.Sprintf("http://%s:%d/server/healthz", getConnectAddress(address), port)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("%s Server is not running\n", icon("❌", "[FAIL]"))
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("%s Server is running but unhealthy\n", icon("⚠️", "[WARN]"))
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	var status map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		fmt.Printf("%s Server is running\n", icon("✅", "[OK]"))
		return nil
	}

	fmt.Printf("%s Server is running and healthy\n", icon("✅", "[OK]"))
	fmt.Printf("   Version:  %v\n", status["version"])
	fmt.Printf("   Mode:     %v\n", status["mode"])
	fmt.Printf("   Uptime:   %v\n", status["uptime"])
	fmt.Printf("   Address:  %s:%d\n", getDisplayAddress(address), port)

	return nil
}

// getConnectAddress returns the address the health check should dial.
// A wildcard bind address is not connectable, so it becomes the loopback
// interface rather than a name that may not resolve inside a container.
func getConnectAddress(address string) string {
	switch address {
	case "", "0.0.0.0", "[::]", "::":
		return "127.0.0.1"
	}
	return address
}

// getDisplayAddress turns a wildcard bind address into something reachable and
// printable, preferring the host name when one is configured.
func getDisplayAddress(address string) string {
	if address == "0.0.0.0" || address == "127.0.0.1" || address == "localhost" || address == "[::]" || address == "" {
		hostname, err := os.Hostname()
		if err == nil && hostname != "" && hostname != "localhost" {
			return hostname
		}
		return "localhost"
	}
	return address
}
