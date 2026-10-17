package migration

import (
	"context"
	"fmt"
	"io"
)

const (
	CommandApply  string = "apply"
	CommandStatus string = "status"
)

// MakeDispatch creates a Dispatch that runs the command named by args[0], defaulting to
// CommandApply when args is empty, through the injected apply and runStatus. Unknown commands or
// extra arguments return ErrUnknownCommand instead of running anything.
func MakeDispatch(apply Apply, runStatus StatusRunner) Dispatch {
	return func(ctx context.Context, args []string) error {
		if len(args) > 1 {
			return ErrUnknownCommand
		}

		command := CommandApply
		if len(args) == 1 {
			command = args[0]
		}

		switch command {
		case CommandApply:
			err := apply(ctx)
			if err != nil {
				return fmt.Errorf("%w: %w", ErrFailedToRunApply, err)
			}
		case CommandStatus:
			err := runStatus(ctx)
			if err != nil {
				return fmt.Errorf("%w: %w", ErrFailedToRunStatus, err)
			}
		default:
			return ErrUnknownCommand
		}

		return nil
	}
}

// MakeStatusRunner creates a StatusRunner that writes one "applied <file>" or "pending <file>" line
// per migration file to out.
func MakeStatusRunner(status Status, out io.Writer) StatusRunner {
	return func(ctx context.Context) error {
		records, err := status(ctx)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToGetStatus, err)
		}

		for _, record := range records {
			state := "pending"
			if record.Applied {
				state = "applied"
			}

			_, err = fmt.Fprintf(out, "%s %s\n", state, record.Name)
			if err != nil {
				return fmt.Errorf("%w: %w", ErrFailedToWriteStatus, err)
			}
		}

		return nil
	}
}
