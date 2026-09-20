package migration

import "context"

// MockListFiles returns a ListFiles that always returns the given names and error.
func MockListFiles(names []string, err error) ListFiles {
	return func() ([]string, error) {
		return names, err
	}
}

// MockTableExists returns a TableExists that always returns the given value and error.
func MockTableExists(exists bool, err error) TableExists {
	return func(context.Context) (bool, error) {
		return exists, err
	}
}

// MockAppliedNames returns an AppliedNames that always returns the given names and error.
func MockAppliedNames(names []string, err error) AppliedNames {
	return func(context.Context) ([]string, error) {
		return names, err
	}
}

// MockStatus returns a Status that always returns the given records and error.
func MockStatus(records []Record, err error) Status {
	return func(context.Context) ([]Record, error) {
		return records, err
	}
}

// MockApply returns an Apply that always returns err.
func MockApply(err error) Apply {
	return func(context.Context) error {
		return err
	}
}
