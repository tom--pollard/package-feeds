package scheduler

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ossf/package-feeds/feeds"
)

// Scheduler is a registry of feeds that should be run on a schedule.
type Scheduler struct {
	Registry map[string]feeds.ScheduledFeed
}

// New returns a new Scheduler with configured feeds registered.
func New(feedsMap map[string]feeds.ScheduledFeed) *Scheduler {
	return &Scheduler{
		Registry: feedsMap,
	}
}

type pollResult struct {
	name     string
	feed     feeds.ScheduledFeed
	packages []*feeds.Package
	err      error
}

// Poll fetches the latest packages from each registered feed, or a specific set.
func (s *Scheduler) Poll(cutoff time.Time, feedsToPoll []string) ([]*feeds.Package, []error) {
	results := make(chan pollResult)
	errs := []error{}
	packages := []*feeds.Package{}
	log.Printf("Value of feedstoPoll inside scheduler %s", feedsToPoll[0])
	for _, feedName := range feedsToPoll {
		log.Printf("Calling go func on %s", feedName)
		go func(name string, feed feeds.ScheduledFeed) {
			result := pollResult{
				name: name,
				feed: feed,
			}
			result.packages, result.err = feed.Latest(cutoff)
			results <- result
		}(feedName, s.Registry[feedName])
	}

	for i := 0; i < len(feedsToPoll); i++ {
		result := <-results

		logger := log.WithField("feed", result.name)
		if result.err != nil {
			logger.WithError(result.err).Error("error fetching packages")
			errs = append(errs, result.err)
			continue
		}
		for _, pkg := range result.packages {
			log.WithFields(log.Fields{
				"feed":    result.name,
				"name":    pkg.Name,
				"version": pkg.Version,
			}).Print("Processing Package")
		}
		packages = append(packages, result.packages...)
		logger.WithField("num_processed", len(result.packages)).Print("processed packages")
	}
	return packages, errs
}
