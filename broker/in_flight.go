package broker

import "sync"

// waitInFlight closes done once every delivery tracked by inFlight has finished.
func waitInFlight(inFlight *sync.WaitGroup, done chan<- struct{}) {
	inFlight.Wait()
	close(done)
}
