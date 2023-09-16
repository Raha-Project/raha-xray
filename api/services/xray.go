package services

import (
	"encoding/json"
	"errors"
	"raha-xray/database/model"
	"raha-xray/logger"
	"raha-xray/util/common"
	"raha-xray/util/json_util"
	"raha-xray/xray"
	"sync"
)

var p *xray.Process
var lock sync.Mutex

type XrayService struct {
	InboundService
	OutboundService
	SettingService
	RuleService
	xray.XrayAPI
}

func (s *XrayService) IsXrayRunning() bool {
	if p == nil {
		return false
	}
	return p.IsRunning()
}

func (s *XrayService) GetXrayVersion() string {
	return p.GetVersion()
}

func (s *XrayService) GetXrayConfig() (*xray.Config, error) {
	xrayConfig, err := s.SettingService.GetXrayDefault()
	if err != nil {
		logger.Error("Error in loading config")
		return nil, err
	}

	inboundConfigs, err := s.getInbounds()
	if err != nil {
		return nil, err
	}
	outboundConfigs, err := s.getOutbounds()
	if err != nil {
		return nil, err
	}
	routingConfigs, err := s.getRoutes(xrayConfig.RoutingConfig)
	if err != nil {
		return nil, err
	}
	xrayConfig.InboundConfigs = append(xrayConfig.InboundConfigs, inboundConfigs...)
	xrayConfig.OutboundConfigs = append(xrayConfig.OutboundConfigs, outboundConfigs...)
	xrayConfig.RoutingConfig = routingConfigs
	return xrayConfig, nil
}

func (s *XrayService) getInbounds() ([]xray.InboundConfig, error) {
	inboundConfigs, err := s.InboundService.GetXrayInboundConfigs()
	if err != nil {
		return nil, err
	}
	return inboundConfigs, nil
}

func (s *XrayService) getOutbounds() ([]json_util.RawMessage, error) {
	var outboundConfigs []json_util.RawMessage
	outbounds, err := s.OutboundService.GetAll()
	if err != nil {
		return nil, err
	}
	for _, outbound := range outbounds {
		outboundJSON, err := s.OutboundService.GetOutboundConfig(outbound)
		if err != nil {
			return nil, err
		}
		outboundConfigs = append(outboundConfigs, *outboundJSON)
	}
	return outboundConfigs, nil
}

func (s *XrayService) getRoutes(routes json_util.RawMessage) (json_util.RawMessage, error) {
	var routeConfig map[string]json_util.RawMessage
	newRouteConfig := make(map[string]json_util.RawMessage)
	err := json.Unmarshal(routes, &routeConfig)
	if err != nil {
		logger.Error("here", string(routes))
		return nil, err
	}
	if common.NonEmptyValue(string(routeConfig["domainStrategy"])) {
		newRouteConfig["domainStrategy"] = routeConfig["domainStrategy"]
	}
	if common.NonEmptyValue(string(routeConfig["domainMatcher"])) {
		newRouteConfig["domainMatcher"] = routeConfig["domainMatcher"]
	}
	if common.NonEmptyValue(string(routeConfig["balancers"])) {
		newRouteConfig["balancers"] = routeConfig["balancers"]
	}
	var rulesConfig []map[string]interface{}
	err = json.Unmarshal([]byte(routeConfig["rules"]), &rulesConfig)
	if err != nil {
		return nil, err
	}
	rules, err := s.RuleService.GetAll()
	if err != nil {
		return nil, err
	}
	for _, rule := range rules {
		ruleConfig := make(map[string]interface{})
		if common.NonEmptyValue(rule.DomainMatcher) {
			ruleConfig["domainMatcher"] = rule.DomainMatcher
		}
		if common.NonEmptyValue(rule.Type) {
			ruleConfig["type"] = rule.Type
		}
		if common.NonEmptyValue(rule.Domain) {
			var domain []string
			json.Unmarshal([]byte(rule.Domain), &domain)
			ruleConfig["domain"] = domain
		}
		if common.NonEmptyValue(rule.Ip) {
			var ip []string
			json.Unmarshal([]byte(rule.Ip), &ip)
			ruleConfig["ip"] = ip
		}
		if common.NonEmptyValue(rule.Port) {
			ruleConfig["port"] = rule.Port
		}
		if common.NonEmptyValue(rule.SourcePort) {
			ruleConfig["sourcePort"] = rule.SourcePort
		}
		if common.NonEmptyValue(rule.Network) {
			ruleConfig["network"] = rule.Network
		}
		if common.NonEmptyValue(rule.Source) {
			var source []string
			json.Unmarshal([]byte(rule.Source), &source)
			ruleConfig["source"] = source
		}
		if common.NonEmptyValue(rule.User) {
			var user []string
			json.Unmarshal([]byte(rule.User), &user)
			ruleConfig["user"] = user
		}
		if common.NonEmptyValue(rule.InboundTag) {
			var inboundTag []string
			json.Unmarshal([]byte(rule.InboundTag), &inboundTag)
			ruleConfig["inboundTag"] = inboundTag
		}
		if common.NonEmptyValue(rule.Protocol) {
			var protocol []string
			json.Unmarshal([]byte(rule.Protocol), &protocol)
			ruleConfig["protocol"] = protocol
		}
		if common.NonEmptyValue(string(rule.Attrs)) {
			var attrs interface{}
			json.Unmarshal([]byte(rule.Attrs), &attrs)
			ruleConfig["attrs"] = attrs
		}
		if common.NonEmptyValue(rule.OutboundTag) {
			ruleConfig["outboundTag"] = rule.OutboundTag
		}
		if common.NonEmptyValue(rule.BalancerTag) {
			ruleConfig["balancerTag"] = rule.BalancerTag
		}
		rulesConfig = append(rulesConfig, ruleConfig)
	}

	rulesJSON, err := json.Marshal(rulesConfig)
	if err != nil {
		return nil, err
	}
	newRouteConfig["rules"] = rulesJSON
	newRouteJSON, err := json.Marshal(newRouteConfig)
	if err != nil {
		return nil, err
	}

	return newRouteJSON, nil
}

func (s *XrayService) GetXrayTraffic() ([]*model.Traffic, error) {
	if !s.IsXrayRunning() {
		return nil, errors.New("xray is not running")
	}
	s.XrayAPI.Init(p.GetAPIServer())
	defer s.XrayAPI.Close()
	return s.XrayAPI.GetTraffic(true)
}

func (s *XrayService) StartProcess() error {
	xrayConfig, err := s.GetXrayConfig()
	if err != nil {
		return err
	}
	p = xray.NewProcess(xrayConfig)
	return p.Start()
}

func (s *XrayService) SetXrayVersion(version string) error {
	return p.SetVersion(version)
}

func (s *XrayService) RestartXray() error {
	lock.Lock()
	defer lock.Unlock()
	logger.Debug("restart xray")
	return p.Restart()
}

func (s *XrayService) StopXray() error {
	lock.Lock()
	defer lock.Unlock()
	logger.Debug("stop xray")
	return p.Stop()
}

func (s *XrayService) WriteConfigFile(needRestart bool) {
	config, err := s.GetXrayConfig()
	if err != nil {
		logger.Error("Error in getting all configs: ", err)
	}
	err, skipRestart := p.WriteConfigFile(config)
	if err != nil {
		logger.Error("Error in writing config file: ", err)
	}
	if !skipRestart && needRestart {
		logger.Debug("Config file saved.")
		s.RestartXray()
	} else {
		logger.Debug("Config saved! No need to restart xray.")
	}
}
