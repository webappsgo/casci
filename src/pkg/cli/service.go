package cli

import "fmt"

func HandleService(command string) error {
switch command {
case "--help":
showServiceHelp()
return nil
case "start":
return fmt.Errorf("use 'casci' without --service to start")
case "stop", "restart", "reload":
fmt.Printf("⚠️  Service %s not yet implemented\n", command)
return nil
case "--install", "--uninstall", "--disable":
fmt.Printf("⚠️  Service %s not yet implemented\n", command)
return nil
default:
return fmt.Errorf("unknown service command: %s", command)
}
}

func showServiceHelp() {
fmt.Print(`
Service Management:
  start          Start server (just run 'casci')
  stop           Stop the service
  restart        Restart the service
  reload         Reload configuration
  --install      Install as system service
  --uninstall    Remove system service
  --disable      Disable service
`)
}
