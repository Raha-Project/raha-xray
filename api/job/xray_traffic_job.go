package job

import (
	"raha-xray/api/services"
	"raha-xray/logger"
)

type XrayTrafficJob struct {
	services.XrayService
	services.TrafficService
}

func NewXrayTrafficJob() *XrayTrafficJob {
	return new(XrayTrafficJob)
}

func (j *XrayTrafficJob) Run() {
	traffics, err := j.XrayService.GetXrayTraffic()
	if err != nil {
		logger.Warning("get xray traffic failed:", err)
		return
	}
	err, needRestart := j.TrafficService.AddTraffic(traffics)
	if err != nil {
		logger.Warning("add traffic failed:", err)
	}
	if needRestart {
		j.XrayService.RestartXray()
	}
}
