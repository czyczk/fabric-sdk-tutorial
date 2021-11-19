package main

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/cmd/timestampstat/calc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func main() {
	config, err := loadConfig("cmd/timestampstat/tasks.yaml")
	if err != nil {
		log.Fatal(err)
	}

	for i, task := range config.Tasks {
		overallConsumption, avgConsumption, err := calc.CalcTimeConsumptions(task.Before, task.After)
		if err != nil {
			log.Fatal(errors.Wrapf(err, "failed on task #%v", i))
		}

		log.Infof("Task #%v-Overall consumption: %v", i, overallConsumption)
		log.Infof("Task #%v-Average consumption: %v", i, avgConsumption)
	}
}
