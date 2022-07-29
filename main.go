package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	tele "github.com/Henry96Markle/telebot"
	"github.com/joho/godotenv"
)

type Configuration struct {
	BotToken         string `json:"bot_token"`
	OwnerTelegramID  int64  `json:"owner_telegram_id"`
	ConnectionString string `json:"connection_string"`
	LogChannelID     int64  `json:"log_channel_id"`
	LoggingToChannel bool   `json:"logging_to_channel"`
}

const (
	DATABASE_NAME   = "telegram"
	COLLECTION_NAME = "user-records"
)

var (
	Config *Configuration

	Bot *tele.Bot

	Data *Database

	TermSig chan os.Signal

	polling = false
)

func ChanLog(input string) {
	if Bot != nil && Config.LoggingToChannel {

		chat, err := Bot.ChatByID(Config.LogChannelID)
		if err == nil {
			Bot.Send(
				chat,
				input,
				tele.ModeHTML,
			)
		} else {
			log.Printf("Error: %v\n", err)
		}
	}
}

func ChanLogf(format string, a ...any) {
	if Bot != nil && Config.LoggingToChannel {
		if Bot != nil && Config.LoggingToChannel {

			chat, err := Bot.ChatByID(Config.LogChannelID)
			if err == nil {
				Bot.Send(
					chat,
					fmt.Sprintf(format, a...),
					tele.ModeHTML,
				)
			} else {
				log.Printf("Error")
			}
		}
	}
}

func init() {
	TermSig = make(chan os.Signal, 1)
	signal.Notify(TermSig, syscall.SIGINT, syscall.SIGTERM)

	println("Initializing..")

	flag.BoolVar(&polling, "poll", false, "set the bot to polling mode")
	flag.Parse()

	// Get configuration

	env_err := godotenv.Load("config.env")

	if env_err != nil {
		log.Printf("error loading configuration: %v\n", env_err)
	}

	owner_env, ok1 := os.LookupEnv("OWNER")
	token_env, ok2 := os.LookupEnv("TOKEN")
	connection_string_env, ok3 := os.LookupEnv("CONNECTION_STRING")
	logging_to_channel_env, ok4 := os.LookupEnv("LOGGING_TO_CHAT")
	log_channel_id_env, ok5 := os.LookupEnv("LOG_CHAT_ID")
	port_env, ok6 := os.LookupEnv("PORT")

	if !ok6 {
		port_env = "80"
	}

	if !(ok1 && ok2 && ok3 && ok4) {
		log.Fatalf("FATAL: unable to acquire evironment variable(s): "+
			"OWNER: %t, TOKEN: %t, CONNECTION_STRING: %t, LOGGING_TO_CHAT: %t\n", ok1, ok2, ok3, ok4)
	}

	if ok4 && logging_to_channel_env == "true" && !ok5 {
		log.Fatalf("FATAL: specified to log into a channel, but was not given an ID for one\n")
	}

	owner_id, owner_err := strconv.ParseInt(owner_env, 0, 64)

	chan_id, chan_err := strconv.ParseInt(log_channel_id_env, 0, 64)

	if owner_err != nil {
		log.Fatalf("FATAL: error parsing owner ID: %v\n", owner_err)
	}

	if chan_err != nil {
		log.Fatalf("FATAL: error parsing log channel ID: %v\n", chan_err)
	}

	doLog, bool_err := strconv.ParseBool(os.Getenv("LOGGING_TO_CHAT"))

	if bool_err != nil {
		log.Fatalf("FATAL: failed to parse bool: %v\n", bool_err)
	}

	Config = &Configuration{
		OwnerTelegramID:  owner_id,
		BotToken:         token_env,
		ConnectionString: connection_string_env,
		LoggingToChannel: doLog,
		LogChannelID:     chan_id,
	}

	// Connect to database

	d, d_err := NewDatabase(Config.ConnectionString, DATABASE_NAME, COLLECTION_NAME)

	if d_err != nil {
		panic(fmt.Errorf("error when connectiong to database: %w", d_err))
	}

	Data = d

	// Initialize bot

	var pref tele.Settings

	if polling {
		pref = tele.Settings{
			Token: Config.BotToken,
			Poller: &tele.LongPoller{
				Timeout: 60 * time.Second,
			},
			Verbose: true,
		}

	} else {
		pref = tele.Settings{
			Token: Config.BotToken,
			Poller: &tele.Webhook{
				Endpoint:       &tele.WebhookEndpoint{PublicURL: "https://botone-bot.herokuapp.com/"},
				AllowedUpdates: []string{"message", "callback_query", "inline_query"},
				Listen:         ":" + port_env,
			},
			Verbose: true,
		}
	}

	b, b_err := tele.NewBot(pref)

	if b_err != nil {
		log.Fatalf("FATAL: error creating initializing bot: %v\n", b_err)
		return
	}

	Bot = b

	Bot.Use(func(hf tele.HandlerFunc) tele.HandlerFunc {
		return func(ctx tele.Context) error {
			toCheck := ""

			if ctx.Query() != nil {
				toCheck = tele.OnQuery
			} else if ctx.Callback() != nil {
				toCheck = ctx.Callback().Unique
			} else {
				toCheck = strings.TrimLeft(strings.Split(ctx.Text(), " ")[0], "/")
			}

			usr, err := Data.FindByID(ctx.Sender().ID)
			per, ok := Permissions[toCheck]

			if ctx.Sender().ID == Config.OwnerTelegramID || err == nil && ok && usr.Permission >= per {
				return hf(ctx)
			} else {
				if ctx.Callback() != nil {
					return ctx.Respond(&tele.CallbackResponse{Text: "You're unauthorized to perform this action"})
				} else {
					return nil
				}
			}
		}
	})

	Bot.Handle(tele.OnQuery, QueryHandler) // <- Not working

	Bot.Handle("/start", func(ctx tele.Context) error {
		handler, ok := CommandMap[ctx.Message().Payload]

		if ok {
			ctx.Message().Payload = ""
			return handler(ctx)
		}

		return nil
	})

	Bot.Handle("/"+CMD_SET, SetHandler)
	Bot.Handle("/"+CMD_REG, RegHandler)
	Bot.Handle("/"+CMD_PERM, PermHandler)
	Bot.Handle("/"+CMD_HELP, HelpHandler)
	Bot.Handle("/"+CMD_ALIAS, AliasHandler)
	Bot.Handle("/"+CMD_UNREG, UnregHandler)
	Bot.Handle("/"+CMD_RECALL, RecallHandler)
	Bot.Handle("/"+CMD_RECORD, RecordHandler)
	Bot.Handle("/"+CMD_CREDITS, CreditsHandler)

	Bot.Handle(SetPermBtn, SetPermBtnHandler)
	Bot.Handle(SetHelpBtn, SetHelpBtnHandler)
	Bot.Handle(RegHelpBtn, RegHelpBtnHandler)
	Bot.Handle(PermHelpBtn, PermHelpBtnHandler)
	Bot.Handle(UnregHelpBtn, UnregHelpBtnHandler)
	Bot.Handle(AliasHelpBtn, AliasHelpBtnHandler)
	Bot.Handle(RecordHelpBtn, RecordHelpBtnHandler)
	Bot.Handle(RecallHelpBtn, RecallHelpBtnHandler)
	Bot.Handle(BackToHelpBtn, BackToHelpBtnHandler)
	Bot.Handle(DeleteEntryBtn, DeleteEntryBtnHandler)
	Bot.Handle(UploadResultBtn, UploadResultBtnHandler)
	Bot.Handle(ConfirmOperatorBtn, ConfirmOperatorBtnHandler)
	Bot.Handle(CancelOperatorConfirmationBtn, CancelOperatorConfirmationBtnHandler)

	Bot.OnError = func(err error, ctx tele.Context) {
		ChanLogf("Error: %v\n", err)
	}
}

func main() {
	var group sync.WaitGroup

	// Start bot

	ChanLog("Starting bot..")
	group.Add(1)
	go func(group *sync.WaitGroup) {
		Bot.Start()

		group.Done()
	}(&group)

	log_term := make(chan bool, 2)

	// group.Add(1)
	// go func(group *sync.WaitGroup, channel <-chan bool) {
	// 	ticker := time.NewTicker(20 * time.Minute)

	// loop:
	// 	for {
	// 		select {
	// 		case <-ticker.C:
	// 			log.Println("listening..")
	// 			ChanLog("Still online.")
	// 		case <-channel:
	// 			break loop
	// 		}
	// 	}

	// 	group.Done()
	// }(&group, log_term)

	ChanLog("Bot is running.")
	if polling {
		s := ""

		fmt.Print("enter any key to terminate")
		fmt.Scanf("%s\n", &s)
	} else {
		<-TermSig
		Bot.RemoveWebhook()
	}

	ChanLog("Shutting down..")
	log.Println("terminating bot..")

	Bot.Stop()
	Bot.Close()

	log.Println("disconnecting..")

	Data.Disconnect()

	log_term <- true
	group.Wait()

	log.Println("Program has ended.")
}
