package lifecycle

import "context"

func MockServe(err error) Serve {
	return func() error {
		return err
	}
}

func MockServeUntilSignal(done <-chan struct{}, err error) Serve {
	return func() error {
		<-done
		return err
	}
}

func MockShutdown(err error) Shutdown {
	return func(context.Context) error {
		return err
	}
}

func MockShutdownRespectingContext() Shutdown {
	return func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}
}

func MockCloseDatabase(calls *int) CloseDatabase {
	return func() {
		*calls++
	}
}

func MockBlockingCloseDatabase(done <-chan struct{}) CloseDatabase {
	return func() {
		<-done
	}
}

func MockRecordingShutdown(order *[]string, err error) Shutdown {
	return func(context.Context) error {
		*order = append(*order, "shutdown")
		return err
	}
}

func MockRecordingCloseDatabase(order *[]string) CloseDatabase {
	return func() {
		*order = append(*order, "close")
	}
}
