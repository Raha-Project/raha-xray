package xray

import (
	"context"
	"encoding/json"
	"raha-xray/database/model"
	"raha-xray/logger"
	"raha-xray/util/common"
	"regexp"
	"time"

	"github.com/xtls/xray-core/app/proxyman/command"
	statsService "github.com/xtls/xray-core/app/stats/command"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/infra/conf"
	"github.com/xtls/xray-core/proxy/shadowsocks"
	"github.com/xtls/xray-core/proxy/shadowsocks_2022"
	"github.com/xtls/xray-core/proxy/trojan"
	"github.com/xtls/xray-core/proxy/vless"
	"github.com/xtls/xray-core/proxy/vmess"
	"google.golang.org/grpc"
)

type XrayAPI struct {
	HandlerServiceClient *command.HandlerServiceClient
	StatsServiceClient   *statsService.StatsServiceClient
	grpcClient           *grpc.ClientConn
	isConnected          bool
}

func (x *XrayAPI) Init(apiServer string) (err error) {
	if len(apiServer) == 0 {
		return common.NewError("wrong xray api server:", apiServer)
	}
	x.grpcClient, err = grpc.Dial(apiServer, grpc.WithInsecure())
	if err != nil {
		return err
	}
	x.isConnected = true

	hsClient := command.NewHandlerServiceClient(x.grpcClient)
	ssClient := statsService.NewStatsServiceClient(x.grpcClient)

	x.HandlerServiceClient = &hsClient
	x.StatsServiceClient = &ssClient

	return
}

func (x *XrayAPI) Close() {
	x.grpcClient.Close()
	x.HandlerServiceClient = nil
	x.StatsServiceClient = nil
	x.isConnected = false
}

func (x *XrayAPI) AddInbound(inbound []byte) error {
	client := *x.HandlerServiceClient

	conf := new(conf.InboundDetourConfig)
	err := json.Unmarshal(inbound, conf)
	if err != nil {
		logger.Debug("Failed to unmarshal inbound:", err)
		return err
	}
	config, err := conf.Build()
	if err != nil {
		logger.Debug("Failed to build inbound Detur:", err)
		return err
	}
	inboundConfig := command.AddInboundRequest{Inbound: config}

	_, err = client.AddInbound(context.Background(), &inboundConfig)

	return err
}

func (x *XrayAPI) DelInbound(tag string) error {
	client := *x.HandlerServiceClient
	_, err := client.RemoveInbound(context.Background(), &command.RemoveInboundRequest{
		Tag: tag,
	})
	return err
}

func (x *XrayAPI) AddOutbound(outbound []byte) error {
	client := *x.HandlerServiceClient

	conf := new(conf.OutboundDetourConfig)
	err := json.Unmarshal(outbound, conf)
	if err != nil {
		logger.Debug("Failed to unmarshal outbound:", err)
		return err
	}
	config, err := conf.Build()
	if err != nil {
		logger.Debug("Failed to build outbound Detur:", err)
		return err
	}
	outboundConfig := command.AddOutboundRequest{Outbound: config}

	_, err = client.AddOutbound(context.Background(), &outboundConfig)

	return err
}

func (x *XrayAPI) DelOutbound(tag string) error {
	client := *x.HandlerServiceClient
	_, err := client.RemoveOutbound(context.Background(), &command.RemoveOutboundRequest{
		Tag: tag,
	})
	return err
}

func (x *XrayAPI) AddUser(Protocol string, inboundTag string, user map[string]interface{}) error {
	var account *serial.TypedMessage
	switch Protocol {
	case "vmess":
		id, ok := user["id"].(string)
		if !ok {
			return common.NewError("Unable to parse vmess client")
		}
		account = serial.ToTypedMessage(&vmess.Account{
			Id: id,
		})
	case "vless":
		id, ok := user["id"].(string)
		if !ok {
			return common.NewError("Unable to parse vless client")
		}
		flow, ok := user["flow"].(string)
		if !ok {
			flow = ""
		}
		account = serial.ToTypedMessage(&vless.Account{
			Id:   id,
			Flow: flow,
		})
	case "trojan":
		pass, ok := user["password"].(string)
		if !ok {
			return common.NewError("Unable to parse trojan client")
		}
		account = serial.ToTypedMessage(&trojan.Account{
			Password: pass,
		})
	case "shadowsocks":
		var ssCipherType shadowsocks.CipherType
		switch user["cipher"].(string) {
		case "aes-128-gcm":
			ssCipherType = shadowsocks.CipherType_AES_128_GCM
		case "aes-256-gcm":
			ssCipherType = shadowsocks.CipherType_AES_256_GCM
		case "chacha20-poly1305", "chacha20-ietf-poly1305":
			ssCipherType = shadowsocks.CipherType_CHACHA20_POLY1305
		case "xchacha20-poly1305", "xchacha20-ietf-poly1305":
			ssCipherType = shadowsocks.CipherType_XCHACHA20_POLY1305
		default:
			ssCipherType = shadowsocks.CipherType_NONE
		}

		pass, ok := user["password"].(string)
		if !ok {
			return common.NewError("Unable to parse shadowsocks client")
		}

		if ssCipherType != shadowsocks.CipherType_NONE {
			account = serial.ToTypedMessage(&shadowsocks.Account{
				Password:   pass,
				CipherType: ssCipherType,
			})
		} else {
			account = serial.ToTypedMessage(&shadowsocks_2022.User{
				Key:   pass,
				Email: user["email"].(string),
			})
		}
	default:
		return nil
	}

	client := *x.HandlerServiceClient

	_, err := client.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag: inboundTag,
		Operation: serial.ToTypedMessage(&command.AddUserOperation{
			User: &protocol.User{
				Email:   user["email"].(string),
				Account: account,
			},
		}),
	})
	return err
}

func (x *XrayAPI) DelUser(inboundTag string, email string) error {
	client := *x.HandlerServiceClient
	_, err := client.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag: inboundTag,
		Operation: serial.ToTypedMessage(&command.RemoveUserOperation{
			Email: email,
		}),
	})
	return err
}

func (x *XrayAPI) GetTraffic(reset bool) ([]*model.Traffic, error) {
	if x.grpcClient == nil {
		return nil, common.NewError("xray api is not initialized")
	}
	var trafficRegex = regexp.MustCompile("(inbound|outbound|user)>>>([^>]+)>>>traffic>>>(downlink|uplink)")

	client := *x.StatsServiceClient
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	request := &statsService.QueryStatsRequest{
		Reset_: reset,
	}
	resp, err := client.QueryStats(ctx, request)
	if err != nil {
		return nil, err
	}

	traffics := make([]*model.Traffic, 0)
	for _, stat := range resp.GetStat() {
		matches := trafficRegex.FindStringSubmatch(stat.Name)
		if len(matches) < 3 || stat.Value == 0 {
			continue
		} else {
			traffics = append(traffics, &model.Traffic{
				DateTime:  uint64(time.Now().Unix()),
				Resource:  matches[1],
				Tag:       matches[2],
				Direction: matches[3] == "downlink",
				Traffic:   uint64(stat.Value),
			})
		}
	}

	return traffics, nil
}

func (x *XrayAPI) GetXrayStats() (*statsService.SysStatsResponse, error) {
	if x.grpcClient == nil {
		return nil, common.NewError("xray api is not initialized")
	}
	client := *x.StatsServiceClient
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	request := &statsService.SysStatsRequest{}
	resp, err := client.GetSysStats(ctx, request)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
