package setup

// Call is a function that CapturePanic runs.
type Call func()

// CapturePanic runs call and returns what it panicked with, or nil when it returned.
func CapturePanic(call Call) (recovered any) {
	defer recoverInto(&recovered)
	call()

	return nil
}

// recoverInto stores in recovered what the deferring function panicked with.
func recoverInto(recovered *any) {
	*recovered = recover()
}
