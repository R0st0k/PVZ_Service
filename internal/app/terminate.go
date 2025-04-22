package app

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func WaitForTerminate(serverStopFuncs ...func(*sync.WaitGroup)) {
	// Terminate
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	var wg sync.WaitGroup
	wg.Add(len(serverStopFuncs))

	for _, serverStopFunc := range serverStopFuncs {
		go serverStopFunc(&wg)
	}
}
