package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/diogonicoleti/pg-activity-watchdog/watchdog"
	_ "github.com/lib/pq"

	"gopkg.in/robfig/cron.v2"
)

var (
	version        = "dev"
	dataSourceName = kingpin.Flag("datasource", "Database connection string").Short('d').
			Default("user=postgres dbname=postgres sslmode=disable").String()
	threshould = kingpin.Flag("threshould", "Threshould to take a snapshot").Short('t').
			Default("30").Int()
)

func main() {
	kingpin.Version(version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infof("Starting PostgreSQL activity watchdog %s", version)
	watchdog := watchdog.NewWatchdog(
		*dataSourceName,
		*threshould,
	)

	c := cron.New()
	if _, err := c.AddFunc("@every 500ms", watchdog.Execute); err != nil {
		log.WithError(err).Fatal("Failed to schedule watchdog")
	}
	c.Start()

	waitInterruptSignal()
	log.Info("Stopping PostgreSQL activity watchdog")
}

func waitInterruptSignal() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		log.Infof("Received signal: %s", sig)
		done <- true
	}()

	<-done
}