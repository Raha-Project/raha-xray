package xray

import (
	"encoding/json"
	"raha-xray/util/json_util"
)

type Config struct {
	LogConfig       json_util.RawMessage   `json:"log"`
	DNSConfig       json_util.RawMessage   `json:"dns"`
	Transport       json_util.RawMessage   `json:"transport"`
	Policy          json_util.RawMessage   `json:"policy"`
	API             json_util.RawMessage   `json:"api"`
	Stats           json_util.RawMessage   `json:"stats"`
	Reverse         json_util.RawMessage   `json:"reverse"`
	FakeDNS         json_util.RawMessage   `json:"fakeDns"`
	InboundConfigs  []InboundConfig        `json:"inbounds"`
	OutboundConfigs []json_util.RawMessage `json:"outbounds"`
	RoutingConfig   json_util.RawMessage   `json:"routing"`
}

type InboundConfig struct {
	Listen         json_util.RawMessage `json:"listen"`
	Port           int                  `json:"port"`
	Protocol       string               `json:"protocol"`
	Settings       json_util.RawMessage `json:"settings"`
	StreamSettings json_util.RawMessage `json:"streamSettings"`
	Tag            string               `json:"tag"`
	Sniffing       json_util.RawMessage `json:"sniffing"`
}

func (c *Config) Equals(other *Config) bool {
	cfg1JSON, _ := json.Marshal(c)
	cfg2JSON, _ := json.Marshal(other)
	return string(cfg1JSON) == string(cfg2JSON)
}
