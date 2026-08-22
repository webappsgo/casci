package cli

import "fmt"

// HandleMaintenance dispatches --maintenance subcommands.
func HandleMaintenance(command, arg string) error {
	switch command {
	case "help", "--help":
		showMaintenanceHelp()
		return nil

	case "backup", "update", "setup":
		return fmt.Errorf("maintenance %s is not implemented yet", command)

	case "restore":
		if arg == "" {
			return fmt.Errorf("backup file required")
		}
		return fmt.Errorf("maintenance restore is not implemented yet")

	case "mode":
		switch arg {
		case "":
			return fmt.Errorf("mode required (production|development|debug)")
		case "production", "development", "debug":
			return fmt.Errorf("maintenance mode change is not implemented yet")
		default:
			return fmt.Errorf("unknown mode: %s", arg)
		}

	default:
		return fmt.Errorf("unknown maintenance command: %s", command)
	}
}

// showMaintenanceHelp prints the --maintenance subcommand reference.
func showMaintenanceHelp() {
	fmt.Print(`Maintenance:
  backup [FILE]               Create a backup archive
  restore FILE                Restore from a backup archive
  update                      Apply pending maintenance updates
  mode {production|development|debug}
                              Change the persisted application mode
  setup                       Run first-time setup
`)
}
