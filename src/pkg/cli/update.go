package cli

import "fmt"

// HandleUpdate dispatches --update subcommands.
func HandleUpdate(command, branch string) error {
	switch command {
	case "help", "--help":
		showUpdateHelp()
		return nil
	case "check", "yes":
		return fmt.Errorf("update %s is not implemented yet", command)
	case "branch":
		if branch == "" {
			return fmt.Errorf("branch required (stable|beta|daily)")
		}
		switch branch {
		case "stable", "beta", "daily":
			return fmt.Errorf("update branch %s is not implemented yet", branch)
		default:
			return fmt.Errorf("unknown update branch: %s", branch)
		}
	default:
		return fmt.Errorf("unknown update command: %s", command)
	}
}

// showUpdateHelp prints the --update subcommand reference.
func showUpdateHelp() {
	fmt.Print(`Update:
  check                       Check whether a newer release exists
  yes                         Download and install the newest release
  branch {stable|beta|daily}  Switch the update channel
`)
}
