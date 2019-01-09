package main

import (
	"log"
	"os"
	"syscall"
)

func checkSignals(sigs chan os.Signal) {
	exitFlag := false
	for !exitFlag {
		sig := <-sigs
		if sig == syscall.SIGINT || sig == syscall.SIGHUP {
			exitFlag = true
			log.Println("Waiting for all goroutines to be completed...")
			wg.Wait()
			log.Println("Exiting process ...")
			os.Exit(1)
		}
	}
}
