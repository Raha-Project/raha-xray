package xray

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"raha-xray/config"
	"raha-xray/util/common"
	"regexp"
	"runtime"
	"strings"
	"time"
)

func GetBinaryName() string {
	os := runtime.GOOS
	if os == "darwin" {
		os = "macos"
	}
	return fmt.Sprintf("xray-%s", os)
}

func GetBinaryPath() string {
	return config.GetXrayFolderPath() + "/" + GetBinaryName()
}

func GetConfigPath() string {
	return config.GetXrayFolderPath() + "/config.json"
}

type Process struct {
	*process
}

func NewProcess(xrayConfig *Config) *Process {
	p := &Process{newProcess(xrayConfig)}
	return p
}

type process struct {
	version   string
	apiServer string

	config *Config
}

func newProcess(config *Config) *process {
	return &process{
		version: "Unknown",
		config:  config,
	}
}

func (p *process) IsRunning() bool {
	// Docker-Run
	if os.Getppid() == 0 {
		_, err := net.DialTimeout("tcp", p.apiServer, 1*time.Second)
		return err == nil
	} else {
		// Non-Docker run
		cmd := exec.Command("pgrep", GetBinaryName())
		output, _ := cmd.CombinedOutput()
		return strings.TrimSpace(string(output)) != ""
	}
}

func (p *process) GetVersion() string {
	return p.version
}

func (p *process) SetVersion(version string) error {
	versionPattern := `^v\d+\.\d+\.\d+$`
	re := regexp.MustCompile(versionPattern)
	if re.MatchString(version) {
		err := os.WriteFile(fmt.Sprintf("%s/version", config.GetXrayFolderPath()), []byte(version), fs.ModePerm)
		if err == nil {
			err = p.Restart()
		}
		return err
	}
	return common.NewError("Wrong version pattern vX.Y.Z")
}

func (p *Process) GetAPIServer() string {
	return p.apiServer
}

func (p *Process) GetConfig() *Config {
	return p.config
}

func (p *process) refreshAPIPort(configs *Config) {
	Env_API := config.GetEnvApi()
	if len(Env_API) > 0 {
		p.apiServer = Env_API
	} else {
		for _, inbound := range configs.InboundConfigs {
			if inbound.Tag == "api" {
				p.apiServer = fmt.Sprintf("%s:%d", inbound.Listen, inbound.Port)
				break
			}
		}
	}
}

func (p *process) refreshVersion() {
	cmd := exec.Command(GetBinaryPath(), "-version")
	data, err := cmd.Output()
	if err != nil {
		p.version = "Unknown"
	} else {
		datas := bytes.Split(data, []byte(" "))
		if len(datas) <= 1 {
			p.version = "Unknown"
		} else {
			p.version = string(datas[1])
		}
	}
}

func (p *process) WriteConfigFile(config *Config) (error, bool) {
	configPath := GetConfigPath()
	p.refreshAPIPort(config)
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return common.NewErrorf("Failure to generate XRAY configuration files: %v", err), false
	}
	// Check if existing file has same config
	if _, err = os.Stat(configPath); err == nil {
		oldConfig, err := os.ReadFile(configPath)
		if err != nil {
			return common.NewErrorf("Read the old configuration file failed: %v", err), false
		}
		if bytes.Equal(oldConfig, data) {
			return nil, true
		}
	}
	err = os.WriteFile(configPath, data, fs.ModePerm)
	if err != nil {
		return common.NewErrorf("Write the configuration file failed: %v", err), false
	}
	return nil, false
}

func (p *process) Start() error {
	err, isSameConfig := p.WriteConfigFile(p.config)
	if err != nil {
		return err
	}

	p.refreshVersion()

	if p.IsRunning() && isSameConfig {
		return errors.New("xray is already running")
	}

	return p.signalXray("restart")
}

func (p *process) Restart() error {
	var err error
	p.refreshVersion()

	err, _ = p.WriteConfigFile(p.config)
	if err != nil {
		return err
	}

	return p.signalXray("restart")
}

func (p *process) Stop() error {
	if !p.IsRunning() {
		return errors.New("xray is not running")
	}

	return p.signalXray("stop")
}

func (p *process) signalXray(signal string) error {
	return os.WriteFile(fmt.Sprintf("%s/xraysignal", config.GetXrayFolderPath()), []byte(signal), fs.ModePerm)
}
