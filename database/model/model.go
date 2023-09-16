package model

type Protocol string

const (
	VMess       Protocol = "vmess"
	VLESS       Protocol = "vless"
	Dokodemo    Protocol = "dokodemo-door"
	Http        Protocol = "http"
	Trojan      Protocol = "trojan"
	Shadowsocks Protocol = "shadowsocks"
)

type User struct {
	Id  uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Key string `json:"key" form:"key"`
}

type Inbound struct {
	Id     uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Name   string `json:"name" form:"name"`
	Enable bool   `json:"enable" form:"enable" gorm:"default:true"`

	// config part
	Listen   string `json:"listen" form:"listen"`
	Port     uint   `json:"port" form:"port"`
	ConfigId uint   `gorm:"not null" json:"configId" form:"configId"`
	Config   Config `gorm:"foreignKey:ConfigId;references:Id" json:"config"`
	Tag      string `gorm:"unique" json:"tag" form:"tag"`

	// clients part
	ClientInbounds []ClientInbound `gorm:"foreignKey:InboundId;references:Id" json:"clients"`
}

type Config struct {
	Id             uint     `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Protocol       Protocol `json:"protocol" form:"protocol"`
	Settings       string   `json:"settings" form:"settings"`
	StreamSettings string   `json:"streamSettings" form:"streamSettings"`
	Sniffing       string   `json:"sniffing" form:"sniffing"`
	ClientSettings string   `json:"clientSettings" form:"clientSettings"`
}

type Client struct {
	Id     uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Name   string `json:"name" form:"name" gorm:"unique"`
	Enable bool   `json:"enable" form:"enable" gorm:"default:true"`
	Quota  uint64 `json:"quota" form:"quota" gorm:"default:0"`
	Expiry uint64 `json:"expiry" form:"expiry" gorm:"default:0"`
	Reset  uint   `json:"reset" from:"reset" gorm:"default:0"`
	Once   uint   `json:"once" from:"once" gorm:"default:0"`
	Up     uint64 `json:"up" form:"up" gorm:"default:0"`
	Down   uint64 `json:"down" form:"down" gorm:"default:0"`
	Remark string `json:"remark" form:"remark"`

	// inbounds part
	ClientInbounds []ClientInbound `gorm:"foreignKey:ClientId;references:Id" json:"inbounds"`
}

type ClientInbound struct {
	Id        uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	InboundId uint   `json:"inboundId" form:"inboundId"`
	ClientId  uint   `json:"clientId" form:"clientId"`
	Config    string `json:"config" form:"config"`
}

type Outbound struct {
	Id             uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	SendThrough    string `json:"sendThrough" form:"sendThrough"`
	Protocol       string `json:"protocol" form:"protocol"`
	Settings       string `json:"settings" form:"settings"`
	Tag            string `gorm:"unique" json:"tag" form:"tag"`
	StreamSettings string `json:"streamSettings" form:"streamSettings"`
	ProxySettings  string `json:"proxySettings" form:"proxySettings"`
	Mux            string `json:"mux" form:"mux"`
}

type Rule struct {
	Id            uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	DomainMatcher string `json:"domainMatcher" form:"domainMatcher"`
	Type          string `json:"type" form:"type"`
	Domain        string `json:"domain" form:"domain"`
	Ip            string `json:"ip" form:"ip"`
	Port          string `json:"port" form:"port"`
	SourcePort    string `json:"sourcePort" form:"sourcePort"`
	Network       string `json:"network" form:"network"`
	Source        string `json:"source" form:"source"`
	User          string `json:"user" form:"user"`
	InboundTag    string `json:"inboundTag" form:"inboundTag"`
	Protocol      string `json:"protocol" form:"protocol"`
	Attrs         string `json:"attrs" form:"attrs"`
	OutboundTag   string `json:"outboundTag" form:"outboundTag"`
	BalancerTag   string `json:"balancerTag" form:"balancerTag"`
}

type Traffic struct {
	Id        uint64 `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	DateTime  uint64 `json:"dateTime" form:"dateTime"`
	Resource  string `json:"resource" form:"resource"`
	Tag       string `json:"tag" form:"tag"`
	Direction bool   `json:"direction" form:"direction"`
	Traffic   uint64 `json:"traffic" form:"traffic"`
}

type Setting struct {
	Id    int    `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Key   string `json:"key" form:"key"`
	Value string `json:"value" form:"value"`
}
