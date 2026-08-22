package cli

import "fmt"

// HandleService dispatches --service subcommands.
func HandleService(command string) error {
	switch command {
	case "help", "--help":
		showServiceHelp()
		return nil
	case "start":
		return fmt.Errorf("run %s without --service to start the server", BinaryName())
	case "stop", "restart", "reload", "--install", "--uninstall", "--disable":
		return fmt.Errorf("service %s is not implemented yet", command)
	default:
		return fmt.Errorf("unknown service command: %s", command)
	}
}

// showServiceHelp prints the --service subcommand reference.
func showServiceHelp() {
	fmt.Printf(`Service Management:
  start          Start the server (run %s with no flags)
  stop           Stop the service
  restart        Restart the service
  reload         Reload configuration
  --install      Install as a system service
  --uninstall    Remove the system service
  --disable      Disable the service
`, BinaryName())
}
