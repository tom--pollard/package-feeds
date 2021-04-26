package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"

	"github.com/ossf/package-feeds/config"
	"github.com/ossf/package-feeds/feeds/scheduler"
	"github.com/ossf/package-feeds/publisher"
)

// FeedHandler is a handler that fetches new packages from various feeds.
type FeedHandler struct {
	scheduler *scheduler.Scheduler
	pub       publisher.Publisher
	pollRate  time.Duration
	lastPoll  time.Time
}

func (handler *FeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cutoff := handler.getCutoff()
	handler.lastPoll = time.Now().UTC()
	pkgs, errs := handler.scheduler.Poll(cutoff)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Errorf("error polling for new packages: %v", err)
		}
	}
	processed := 0
	for _, pkg := range pkgs {
		processed++
		log.WithFields(log.Fields{
			"name":         pkg.Name,
			"feed":         pkg.Type,
			"created_date": pkg.CreatedDate,
		}).Print("sending package upstream")
		b, err := json.Marshal(pkg)
		if err != nil {
			log.Printf("error marshaling package: %#v", pkg)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := handler.pub.Send(context.Background(), b); err != nil {
			log.Printf("error sending package to upstream publisher %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if len(errs) > 0 {
		http.Error(w, "error polling for packages - see logs for more information", http.StatusInternalServerError)
		return
	}
	_, err := w.Write([]byte(fmt.Sprintf("%d packages processed", processed)))
	if err != nil {
		http.Error(w, "unexpected error during http server write: %w", http.StatusInternalServerError)
	}
}

func (handler FeedHandler) getCutoff() time.Time {
	var cutoff time.Time
	if handler.lastPoll.IsZero() {
		cutoff = time.Now().UTC().Add(-handler.pollRate)
	} else {
		cutoff = handler.lastPoll
	}
	return cutoff
}

func main() {
	configPath, useConfig := os.LookupEnv("PACKAGE_FEEDS_CONFIG_PATH")
	var err error

	var appConfig *config.ScheduledFeedConfig
	if useConfig {
		appConfig, err = config.FromFile(configPath)
		log.Infof("Using config from file: %v", configPath)
	} else {
		appConfig = config.Default()
		log.Info("No config specified, using default configuration")
	}
	if err != nil {
		log.Fatal(err)
	}

	pub, err := appConfig.PubConfig.ToPublisher(context.TODO())
	if err != nil {
		log.Fatal(fmt.Errorf("failed to initialize publisher from config: %w", err))
	}
	log.Infof("using %q publisher", pub.Name())

	feeds, err := appConfig.GetScheduledFeeds()
	log.Infof("watching feeds: %v", strings.Join(appConfig.EnabledFeeds, ", "))
	if err != nil {
		log.Fatal(err)
	}
	sched := scheduler.New(feeds)

	log.Printf("listening on port %v", appConfig.HTTPPort)
	pollRate, err := time.ParseDuration(appConfig.PollRate)
	if err != nil {
		log.Fatal(err)
	}
	handler := &FeedHandler{
		scheduler: sched,
		pub:       pub,
		pollRate:  pollRate,
	}

	if appConfig.Timer {
		cronjob := cron.New()
		crontab := fmt.Sprintf("@every %s", pollRate.String())
		log.Printf("Running a timer %s", crontab)
		err := cronjob.AddFunc(crontab, func() { cronRequest(appConfig.HTTPPort) })
		if err != nil {
			log.Fatal(err)
		}
		cronjob.Start()
	}

	http.Handle("/", handler)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", appConfig.HTTPPort), nil); err != nil {
		log.Fatal(err)
	}
}

func cronRequest(port int) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%v", port))
	if err != nil {
		log.Printf("http request failed: %v", err)
		return
	}
	resp.Body.Close()
}
