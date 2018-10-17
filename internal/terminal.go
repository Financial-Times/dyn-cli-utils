package terminal

import (
	"time"

	"github.com/briandowns/spinner"
)

func WaitWithSpinner(duration time.Duration) {
	s := spinner.New(spinner.CharSets[21], 100*time.Millisecond)
	s.Start()
	time.Sleep(duration)
	s.Stop()
}
