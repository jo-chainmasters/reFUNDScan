package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/fatih/color"
	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)
var (
    // Persistent json responses
    cg CoinGeckoResponse
    vals ValidatorResponse

    config Config

    // Flags
    configpath string
    initconfig bool
)

func init(){
    flag.StringVar(&configpath,"config", ".", "Directory containing your config.toml")
    flag.BoolVar(&initconfig,"init", false, "Creates a predefined config.toml file, if the config path is not set, defaults to the CWD")
    flag.Parse()
    configpath = strings.TrimRight(configpath,"/")
    if initconfig {
        initConfig(configpath) 
        log.Println(color.GreenString("Config file generated at: " + configpath + "/config.toml"))
        os.Exit(1)
    }
    config.parseConfig(configpath)
    // config.showConfig()
}

// Start the telegram bot and listen for messages from the resp channel
func main(){
    interrupt := make(chan os.Signal, 1) 
    signal.Notify(interrupt, os.Interrupt) 

    resp := make(chan string)
    restart := make(chan bool)
    go Connect(resp, restart)
    bot, err := telegram.NewBotAPI(config.API)
    if err != nil {
        log.Fatal(color.RedString("Cannot connect to bot, check your BotKey or internet connection"))
    }
    // bot.Debug = true

    // AutoRefresh coin gecko and validator set data
    go autoRefresh(config.RestCoinGecko,&cg)
    go autoRefresh(config.RestValidators,&vals)

    go func(){
        for {
            select {
            case message := <- resp:
                msg := telegram.NewMessageToChannel(config.ChatID, message)
                msg.ParseMode = telegram.ModeHTML
                msg.DisableWebPagePreview = true
                _, err := bot.Send(msg)
                if err != nil {
                    log.Println(color.YellowString("Could not sent message, check your internet connection or ChatID"))
                }
                log.Println(color.BlueString(message))
            case <- restart:
                log.Println(color.BlueString("Restarting websocket connection in 30 seconds"))
                time.Sleep(time.Second * 30)
                go Connect(resp, restart)
            }
        }
    }()
    select {
    case <- interrupt:
        log.Println(color.RedString("Interrupted"))
        return
    }
}
