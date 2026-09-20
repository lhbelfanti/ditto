package migration

import (
	"context"
	"fmt"
	"io"
)

const (
	// CommandApply runs pending migrations through Apply. It is also the default when no
	// argument is given, preserving a no-argument-apply CLI's prior behavior.
	CommandApply string = "apply"

	// CommandStatus reports applied/pending migration files without mutating the database.
	CommandStatus string = "status"
)

// Dispatch runs the migrations command named by args[0] (defaulting to CommandApply when args is
// empty), using the injected apply and status functions. Unknown commands or extra arguments
// return ErrUnknownCommand instead of running anything.
func Dispatch(ctx context.Context, args []string, apply Apply, status Status, out io.Writer) error {
	if len(args) > 1 {
		return ErrUnknownCommand
	}

	command := CommandApply
	if len(args) == 1 {
		command = args[0]
	}

	switch command {
	case CommandApply:
		return apply(ctx)
	case CommandStatus:
		return runStatus(ctx, status, out)
	default:
		return ErrUnknownCommand
	}
}

func runStatus(ctx context.Context, status Status, out io.Writer) error {
	records, err := status(ctx)
	if err != nil {
		return err
	}

	for _, record := range records {
		state := "pending"
		if record.Applied {
			state = "applied"
		}

		if _, err := fmt.Fprintf(out, "%s %s\n", state, record.Name); err != nil {
			return err
		}
	}

	return nil
}
