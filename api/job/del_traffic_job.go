package job

import (
	"raha-xray/api/services"
	"raha-xray/logger"
)

type DelOldTrafficJob struct {
	services.TrafficService
}

func NewDelTrafficJob() *DelOldTrafficJob {
	return new(DelOldTrafficJob)
}

func (j *DelOldTrafficJob) Run() {
	result := j.TrafficService.DelOldTraffics()
	logger.Debug("Deleted old traffics:", result)
}
