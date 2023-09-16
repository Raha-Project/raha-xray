package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"raha-xray/api"
	"raha-xray/api/global"
	"raha-xray/config"
	"raha-xray/database"
	"raha-xray/database/model"
	"raha-xray/logger"
	"raha-xray/util/random"
	"syscall"
	_ "unsafe"

	"github.com/op/go-logging"
)

func runServer() {
	log.Printf("%v %v", config.GetName(), config.GetVersion())

	switch config.GetLogLevel() {
	case config.Debug:
		logger.InitLogger(logging.DEBUG)
	case config.Info:
		logger.InitLogger(logging.INFO)
	case config.Warn:
		logger.InitLogger(logging.WARNING)
	case config.Error:
		logger.InitLogger(logging.ERROR)
	default:
		log.Fatal("unknown log level:", config.GetLogLevel())
	}

	err := config.LoadSettings()
	if err != nil {
		log.Println("Failed to load app settings", err)
		return
	}

	err = database.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	server := api.NewServer()
	global.SetWebServer(server)
	err = server.Start()
	if err != nil {
		log.Println(err)
		return
	}

	sigCh := make(chan os.Signal, 1)
	// Trap shutdown signals
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGSEGV)
	for {
		sig := <-sigCh

		switch sig {
		case syscall.SIGHUP:
			err := server.Stop()
			if err != nil {
				logger.Warning("stop server err:", err)
			}
			server = api.NewServer()
			global.SetWebServer(server)
			err = server.Start()
			if err != nil {
				log.Println(err)
				return
			}
		default:
			server.Stop()
			return
		}
	}
}

func getTokens(id int) {
	err := config.LoadSettings()
	if err != nil {
		log.Println("Failed to load app settings", err)
		return
	}

	err = database.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	db := database.GetDB()
	users := []model.User{}
	if id == 0 {
		err = db.Model(model.User{}).Find(&users).Error
	} else {
		err = db.Model(model.User{}).Where("id = ?", id).Find(&users).Error
	}
	if err != nil {
		log.Fatal(err)
	}
	if len(users) > 0 {
		println("ID\t\tTOKEN")
		println("--------*----------")
		for _, user := range users {
			println(user.Id, "\t\t", user.Key)
		}
	} else {
		println("No token found!")
	}
}

func addToken() {
	err := config.LoadSettings()
	if err != nil {
		log.Println("Failed to load app settings", err)
		return
	}

	err = database.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	db := database.GetDB()
	user := &model.User{
		Key: random.Seq(32),
	}
	err = db.Create(user).Error
	if err != nil {
		log.Fatal(err)
	}
	println("ID\tTOKEN")
	println("--------*----------")
	println(user.Id, "\t", user.Key)
}

func delToken(id int) {
	err := config.LoadSettings()
	if err != nil {
		log.Println("Failed to load app settings", err)
		return
	}

	err = database.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	db := database.GetDB()
	err = db.Delete(model.User{}, id).Error
	if err != nil {
		log.Fatal(err)
	} else {
		println("Token ", id, " is now deleted.")
	}
}

func main() {
	if len(os.Args) < 2 {
		runServer()
		return
	}

	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "show version")

	tokenCmd := flag.NewFlagSet("token", flag.ExitOnError)
	var id int
	var list bool
	var add bool
	var del int
	tokenCmd.IntVar(&id, "id", 0, "get token by ID")
	tokenCmd.BoolVar(&list, "list", false, "list all tokens")
	tokenCmd.BoolVar(&add, "add", false, "add a token")
	tokenCmd.IntVar(&del, "del", 0, "delete token by ID")

	tokenCmd.Usage = func() {
		println("token usage:")
		println("\ttoken -id <id>\t\tget token by ID")
		println("\ttoken -list\t\tlist all tokens")
		println("\ttoken -add\t\tadd a new token")
		println("\ttoken -del <id>\t\tdelete token by ID")
	}

	oldUsage := flag.Usage
	flag.Usage = func() {
		oldUsage()
		println("  token\ttoken subcommand\n")
		tokenCmd.Usage()
	}

	flag.Parse()
	if showVersion {
		println(config.GetVersion())
		return
	}
	switch os.Args[1] {
	case "token":
		err := tokenCmd.Parse(os.Args[2:])
		if err != nil {
			println(err)
			return
		}
		if id > 0 {
			getTokens(id)
		}
		if list || len(os.Args) == 2 {
			getTokens(0)
		}
		if add {
			addToken()
		}
		if del > 0 {
			delToken(del)
		}
	default:
		flag.Usage()
	}
}
