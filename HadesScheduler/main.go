package main

import (
	"fmt"
	"os"

	"github.com/Mtze/HadesCI/hadesScheduler/docker"
	"github.com/Mtze/HadesCI/shared/payload"
	"github.com/Mtze/HadesCI/shared/queue"
	"github.com/Mtze/HadesCI/shared/utils"

	log "github.com/sirupsen/logrus"
)

var BuildQueue *queue.Queue

type JobScheduler interface {
	ScheduleJob(job payload.QueuePayload) error
}

func main() {
	if is_debug := os.Getenv("DEBUG"); is_debug == "true" {
		log.SetLevel(log.DebugLevel)
		log.Warn("DEBUG MODE ENABLED")
	}

	var cfg utils.RabbitMQConfig
	utils.LoadConfig(&cfg)

	var executorCfg utils.ExecutorConfig
	utils.LoadConfig(&executorCfg)
	log.Debug("Executor config: ", executorCfg)

	var err error
	rabbitmqURL := fmt.Sprintf("amqp://%s:%s@%s/", cfg.User, cfg.Password, cfg.Url)
	log.Debug("Connecting to RabbitMQ: ", rabbitmqURL)
	BuildQueue, err = queue.Init("builds", rabbitmqURL)

	if err != nil {
		log.Panic(err)
	}

	var forever chan struct{}

	var scheduler JobScheduler

	switch executorCfg.Executor {
	// case "k8s":
	// 	log.Info("Started HadesScheduler in Kubernetes mode")
	// 	kube.Init()
	// 	scheduler = kube.Scheduler{}
	case "docker":
		log.Info("Started HadesScheduler in Docker mode")
		scheduler = docker.Scheduler{}
	default:
		log.Fatalf("Invalid executor specified: %s", executorCfg.Executor)
	}

	BuildQueue.Dequeue(scheduler.ScheduleJob)

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}
