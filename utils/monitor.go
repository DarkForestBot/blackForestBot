//Package utils used for monitoring
package utils

import (
	"log"
	"runtime"

	"github.com/robfig/cron"
)

//DummyForLoad just for main load this package
var DummyForLoad int

func init() {
	crontab := cron.New()
	crontab.AddFunc("@every 10s", Stats)
	crontab.Start()
}

//Stats return threads
func Stats() {
	log.Println("Current goroutines:", runtime.NumGoroutine())
}
