package scheduler

import (
	"time"
	"log"
	"os"

	"github.com/go-co-op/gocron/v2"
)

func Start_scheduler() {
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Printf("failed to start scheduler: %v", err)
	}

	j, err := s.NewJob(
		gocron.DurationJob(
			time.Minute * 5,
		),
		gocron.NewTask(
			func () {
				if delete_err := os.RemoveAll("audio"); delete_err != nil {
					log.Printf("failed to delete directory with audio: %v", delete_err)
				}
			},
		),
	)
	if err != nil {
		log.Printf("failed to add delete job: %v", err)
	}

	s.Start()

	log.Printf("delete job with id started: %v", j.ID())
}