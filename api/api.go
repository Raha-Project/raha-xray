package api

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"raha-xray/api/handlers"
	"raha-xray/api/job"
	"raha-xray/api/network"
	"raha-xray/api/services"
	"raha-xray/config"
	"raha-xray/logger"
	"raha-xray/util/common"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

// to keep uptime
var startTime = time.Now()

type Server struct {
	httpServer *http.Server
	listener   net.Listener

	Inbound  *handlers.InboundHandler
	Config   *handlers.ConfigHandler
	Client   *handlers.ClientHandler
	Outbound *handlers.OutboundHandler
	Rule     *handlers.RuleHandler
	Server   *handlers.ServerHandler
	Setting  *handlers.SettingsHandler

	SettingService services.SettingService
	XrayService    services.XrayService

	appSettings *config.Setting
	cron        *cron.Cron

	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Server) initRouter() (*gin.Engine, error) {
	if config.IsDebug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.DefaultWriter = io.Discard
		gin.DefaultErrorWriter = io.Discard
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.Default()

	g := engine.Group(s.appSettings.BasePath)

	s.Inbound = handlers.NewInboundHandler(g)
	s.Config = handlers.NewConfigHandler(g)
	s.Client = handlers.NewClientHandler(g)
	s.Outbound = handlers.NewOutboundHandler(g)
	s.Rule = handlers.NewRuleHandler(g)
	s.Server = handlers.NewServerHandler(g)
	s.Setting = handlers.NewSettingsHandler(g)

	return engine, nil
}

func (s *Server) startTask() {
	err := s.XrayService.StartProcess()
	if err != nil {
		logger.Warning("start xray failed:", err)
	}

	appConfig := config.GetSettings()

	go func() {
		time.Sleep(time.Second * 5)
		// Statistics every 10 seconds, start the delay for 5 seconds for the first time, and staggered with the time to restart xray
		s.cron.AddJob("@every 10s", job.NewXrayTrafficJob())

		if appConfig.TrafficDays != 0 {
			// Daily deleting old traffics
			s.cron.AddJob("@daily", job.NewDelTrafficJob())
		}
	}()
}

func (s *Server) Start() error {
	var err error
	defer func() {
		if err != nil {
			s.Stop()
		}
	}()

	err = s.SettingService.LoadXrayDefaults()
	if err != nil {
		logger.Error("Failed to load default xray config")
		return err
	}

	s.appSettings = s.SettingService.GetSettings()

	loc, err := s.appSettings.GetTimeLocation()
	if err != nil {
		return err
	}
	s.cron = cron.New(cron.WithLocation(loc), cron.WithSeconds())
	s.cron.Start()

	engine, err := s.initRouter()
	if err != nil {
		return err
	}

	certFile := s.appSettings.CertFile
	keyFile := s.appSettings.KeyFile
	listen := s.appSettings.Listen
	port := s.appSettings.Port

	listenAddr := net.JoinHostPort(listen, strconv.Itoa(port))
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	if certFile != "" || keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			listener.Close()
			return err
		}
		c := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		listener = network.NewAutoHttpsListener(listener)
		listener = tls.NewListener(listener, c)
	}

	if certFile != "" || keyFile != "" {
		logger.Info("web server run https on", listener.Addr())
	} else {
		logger.Info("web server run http on", listener.Addr())
	}
	s.listener = listener

	s.startTask()

	s.httpServer = &http.Server{
		Handler: engine,
	}

	go func() {
		s.httpServer.Serve(listener)
	}()

	return nil
}

func (s *Server) Stop() error {
	s.cancel()
	if s.cron != nil {
		s.cron.Stop()
	}

	var err1 error
	var err2 error
	if s.httpServer != nil {
		err1 = s.httpServer.Shutdown(s.ctx)
	}
	if s.listener != nil {
		err2 = s.listener.Close()
	}
	return common.Combine(err1, err2)
}

func (s *Server) GetUptime() uint64 {
	return uint64(time.Since(startTime).Seconds())
}

func (s *Server) GetCtx() context.Context {
	return s.ctx
}

func (s *Server) GetCron() *cron.Cron {
	return s.cron
}
