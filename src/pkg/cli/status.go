package cli

import (
"encoding/json"
"fmt"
"net/http"
"os"
"time"
)

// ShowStatus displays server status
func ShowStatus(address string, port int) error {
url := fmt.Sprintf("http://%s:%d/health", address, port)

client := &http.Client{Timeout: 5 * time.Second}
resp, err := client.Get(url)

if err != nil {
fmt.Println("❌ Server is not running")
fmt.Printf("   Error: %v\n", err)
return nil
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
fmt.Println("⚠️  Server is running but unhealthy")
fmt.Printf("   Status: %d\n", resp.StatusCode)
return nil
}

var status map[string]interface{}
if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
fmt.Println("✅ Server is running")
return nil
}

fmt.Println("✅ Server is running and healthy")
fmt.Printf("   Version:  %v\n", status["version"])
fmt.Printf("   Mode:     %v\n", status["mode"])
fmt.Printf("   Uptime:   %v\n", status["uptime"])
fmt.Printf("   Address:  %s:%d\n", getDisplayAddress(address), port)

return nil
}

func getDisplayAddress(address string) string {
if address == "0.0.0.0" || address == "127.0.0.1" || address == "localhost" || address == "[::]" {
hostname, err := os.Hostname()
if err == nil && hostname != "" && hostname != "localhost" {
return hostname
}
return "localhost"
}
return address
}
