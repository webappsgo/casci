package cli

import (
"flag"
"fmt"
"runtime"
)

// Config holds CLI configuration
type Config struct {
Mode      string
DataDir   string
ConfigDir string
Address   string
Port      int
ShowHelp  bool
ShowVersion bool
ShowStatus bool
Service   string
Maintenance string
MaintenanceArg string
Update string
UpdateBranch string
}

// Parse parses command line arguments
func Parse() (*Config, error) {
cfg := &Config{}

flag.StringVar(&cfg.Mode, "mode", "", "Application mode (production|development)")
flag.StringVar(&cfg.DataDir, "data", "", "Data directory")
flag.StringVar(&cfg.ConfigDir, "config", "", "Config directory")
flag.StringVar(&cfg.Address, "address", "", "Listen address")
flag.IntVar(&cfg.Port, "port", 0, "Listen port")
flag.BoolVar(&cfg.ShowHelp, "help", false, "Show help")
flag.BoolVar(&cfg.ShowVersion, "version", false, "Show version")
flag.BoolVar(&cfg.ShowStatus, "status", false, "Show status")
flag.StringVar(&cfg.Service, "service", "", "Service command")
flag.StringVar(&cfg.Maintenance, "maintenance", "", "Maintenance command")
flag.StringVar(&cfg.Update, "update", "", "Update command")

flag.Usage = ShowHelp
flag.Parse()

if cfg.Update == "branch" && flag.NArg() > 0 {
cfg.UpdateBranch = flag.Arg(0)
}

if cfg.Maintenance != "" && flag.NArg() > 0 {
cfg.MaintenanceArg = flag.Arg(0)
}

return cfg, nil
}

// ShowHelp displays help
func ShowHelp() {
fmt.Print(`
CASCI - CI/CD Application Server

Usage: casci [OPTIONS]

Information (No Privileges Required):
  --help            Show this help
  --version         Show version
  --status          Show status

Server Options:
  --mode {prod|dev} Set mode
  --port {port}     Set port
  --address {addr}  Set address

See documentation for full command list.
`)
}

// ShowVersion displays version
func ShowVersion(v, b, c string) {
fmt.Printf("casci v%s\nBuilt: %s\nCommit: %s\nGo: %s\nOS/Arch: %s/%s\n",
v, b, c, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// ShouldRunServer returns true if server should start
func (c *Config) ShouldRunServer() bool {
return !c.ShowHelp&&!c.ShowVersion&&!c.ShowStatus&&c.Update!="check"
}
