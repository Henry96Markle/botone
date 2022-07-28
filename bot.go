package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tele "github.com/Henry96Markle/telebot"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	CMD_HELP    = "help"
	CMD_REG     = "reg"
	CMD_UNREG   = "unreg"
	CMD_RECORD  = "record"
	CMD_ALIAS   = "alias"
	CMD_RECALL  = "recall"
	CMD_SET     = "set"
	CMD_CREDITS = "credits"
	CMD_PERM    = "perm"

	// Button unique strings

	BTN_RECORD_HELP   = "recordCommandHelpBtn"
	BTN_REG_HELP      = "regCommandHelpBtn"
	BTN_ALIAS_HELP    = "aliasCommandHelpBtn"
	BTN_RECALL_HELP   = "recallCommandHelpBtn"
	BTN_UNREG_HELP    = "unregCommandHelpBtn"
	BTN_BACK_TO_HELP  = "backToHelpMainPageBtn"
	BTN_UPLOAD_RESULT = "uploadResultBtn"
	BTN_SET_HELP      = "setCommandHelpBtn"
	BTN_PERM_HELP     = "permComandHelpBtn"

	BTN_CANCEL_OPERATOR_CONFIRMATION = "cancelBtn"
	Btn_CONFIRM_OPERATOR             = "confirmOperatorBtn"

	BUFF_FILE_PATH = "./b.txt"

	VERSION = "0.0.5"
)

var (
	Commands = []string{
		CMD_HELP, CMD_REG, CMD_RECORD, CMD_ALIAS,
		CMD_RECALL, CMD_UNREG, CMD_SET, CMD_CREDITS,
		CMD_PERM,
	}

	CommandMap = map[string]func(tele.Context) error{
		CMD_HELP:    HelpHandler,
		CMD_ALIAS:   AliasHandler,
		CMD_CREDITS: CreditsHandler,
		CMD_RECALL:  RecallHandler,
		CMD_RECORD:  RecordHandler,
		CMD_REG:     RegHandler,
		CMD_UNREG:   UnregHandler,
		CMD_SET:     SetHandler,
		CMD_PERM:    PermHandler,
	}

	Permissions = map[string]int{
		CMD_SET:     3,
		CMD_ALIAS:   2,
		CMD_RECORD:  2,
		CMD_UNREG:   2,
		CMD_REG:     2,
		CMD_RECALL:  1,
		CMD_HELP:    1,
		CMD_CREDITS: 1,
		CMD_PERM:    1,

		BTN_RECORD_HELP:                  2,
		BTN_ALIAS_HELP:                   2,
		BTN_BACK_TO_HELP:                 1,
		BTN_RECALL_HELP:                  1,
		BTN_REG_HELP:                     2,
		BTN_UNREG_HELP:                   2,
		BTN_UPLOAD_RESULT:                1,
		BTN_SET_HELP:                     3,
		BTN_PERM_HELP:                    1,
		BTN_CANCEL_OPERATOR_CONFIRMATION: 4,
		Btn_CONFIRM_OPERATOR:             4,

		tele.OnQuery: 1,
	}

	CommandSyntax = map[string]string{
		CMD_REG: "Register new users.\n\nSyntax:\n\n/reg <ID/reply-to-message>",

		CMD_RECORD: "Write down what the user did under a certain category.\n\n" +
			"Syntax:\n\n/record <ID/reply-to-message> <category> [note1; note2; note3; ..]\n\nExample:\n\n" +
			"/record 69696969 bans shared a pirated movie; he blamed me for eating his sandwish",

		CMD_RECALL: "Recall information about a person who's registered before.\n\nSyntax:\n\n" +
			"- /recall <ID/reply-to-message>\n\n" +
			"- /recall <name/username> <value>\n\nExamples:\n\n" +
			"/recall 69696969\n/recall name Miles Edgeworth",

		CMD_ALIAS: "Add more IDs, names, or usernames that belong to the same person.\n\nSyntax:\n\n" +
			"/alias <ID/reply-to-message> <add/remove> <id/name/username> <value1>; <value2> ..\n\nExample:\n\n" +
			"/alias 69696969 add name Henry Markle; Steward; Rose Smith",

		CMD_HELP: "Learn each command's syntax by typing /help followed by the name of the command.\n\nSyntax:\n\n" +
			"- /help\n- /help <command>",

		CMD_UNREG: "There are some people you just want to forget.\n" +
			"Unregister and delete them from the database.\n\nSyntax:\n\n" +
			"- /unreg <ID/reply-to-message>",
		CMD_SET: "You can grant some users access to the database. Either read-only or read/write permissions.\n" +
			"The permissions available are:\n\n- [0] none\n- [1] read\n- [2] write\n\n" +
			"Syntax:\n\n/set <ID> <permission>\n\nExamples:\n\n/set 6969669 read\n/set 1070000 write\n/set 69 none",
		CMD_PERM: "View the permission level of registered users.\n\nSyntax:\n\n" +
			"/perm <ID/reply-to-message>",
	}

	StringBuffer = ""

	// Buttons

	CancelOperatorConfirmationBtn = &tele.Btn{
		Unique: BTN_CANCEL_OPERATOR_CONFIRMATION,
		Text:   "Cancel",
	}

	ConfirmOperatorBtn = &tele.Btn{
		Unique: Btn_CONFIRM_OPERATOR,
		Text:   "Confirm",
	}

	SetHelpBtn = &tele.Btn{
		Unique: BTN_SET_HELP,
		Text:   CMD_SET,
	}

	PermHelpBtn = &tele.Btn{
		Unique: BTN_PERM_HELP,
		Text:   CMD_PERM,
	}

	RecordHelpBtn = &tele.Btn{
		Unique: BTN_RECORD_HELP,
		Text:   CMD_RECORD,
	}

	RegHelpBtn = &tele.Btn{
		Unique: BTN_REG_HELP,
		Text:   CMD_REG,
	}

	AliasHelpBtn = &tele.Btn{
		Unique: BTN_ALIAS_HELP,
		Text:   CMD_ALIAS,
	}

	RecallHelpBtn = &tele.Btn{
		Unique: BTN_RECALL_HELP,
		Text:   CMD_RECALL,
	}

	UnregHelpBtn = &tele.Btn{
		Unique: BTN_UNREG_HELP,
		Text:   CMD_UNREG,
	}

	BackToHelpBtn = &tele.Btn{
		Unique: BTN_BACK_TO_HELP,
		Text:   "Go back",
	}

	UploadResultBtn = &tele.Btn{
		Unique: BTN_UPLOAD_RESULT,
		Text:   "Send in a file",
	}

	InviteBtn = func() *tele.Btn {
		return &tele.Btn{
			Unique: "inviteBtn",
			Text:   "PM me",
			URL:    "https://t.me/" + Bot.Me.Username + "?start=help",
		}
	}

	OperatorConfirmationKeyboard = &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{
			{*CancelOperatorConfirmationBtn.Inline(), *ConfirmOperatorBtn.Inline()},
		},
	}

	UploadResultBtnKeyboard = &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{
			{*UploadResultBtn.Inline()},
		},
	}

	HelpMainPageKeyboard = &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{
			{*RecordHelpBtn.Inline(), *AliasHelpBtn.Inline(), *RecallHelpBtn.Inline()},
			{*RegHelpBtn.Inline(), *UnregHelpBtn.Inline()},
			{*SetHelpBtn.Inline(), *PermHelpBtn.Inline()},
		},
	}

	BackToHelpKeyboard = &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{
			{*BackToHelpBtn.Inline()},
		},
	}

	CreditsKeyboard = &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{
			{*tele.Btn{Unique: "creatorBtn", Text: "Botone", URL: "https://github.com/Henry96Markle/botone"}.Inline()},
			{
				*tele.Btn{Unique: "telebotBtn", Text: "Telebot", URL: "https://github.com/tucnak/telebot"}.Inline(),
				*tele.Btn{Unique: "mongodbBtn", Text: "MongoDB", URL: "https://github.com/mongodb/mongo-go-driver"}.Inline(),
			},
		},
	}
)

func CreditsHandler(ctx tele.Context) error {
	return ctx.Reply(
		fmt.Sprintf(
			"<b>Botone v%s</b>\n\nCreator: <b>Henry Markle</b>\n"+
				"Powered by: <b>Go v1.18</b> & <b>MongoDB</b>\n\n"+
				"Code dependencies:\n\n\t"+
				"<b>Telebot</b> by <a href=\"https://github.com/tucnak\">tucnak</a>\n\t"+
				"<b>Mongo Go Driver</b> by <a href=\"https://www.mongodb.com/\">MongoDB</a>",
			VERSION,
		),
		CreditsKeyboard, tele.ModeHTML, tele.NoPreview,
	)
}

// Syntax:
//
//	- /help
//	- /help <command>
func HelpHandler(ctx tele.Context) error {
	if len(ctx.Args()) == 0 {
		if ctx.Chat().ID == ctx.Sender().ID {
			return ctx.Reply(
				"Don't you hate it when poeple constantly change their names, usernames, and even their Telegram IDs, "+
					"and then you tend to forget who they were and what they did?\n"+
					"With Botone, you can keep track of their identities and record "+
					"their most significant actions, so you don't have to worry about forgetting and feeling like "+
					"everyone on telegram is the same person.\n\n"+
					"Click on the buttons below, to learn each command.",
				HelpMainPageKeyboard, tele.ModeHTML,
			)
		} else {
			return ctx.Reply("PM me to learn how to use the bot.", &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					{*InviteBtn().Inline()},
				},
			})
		}
	} else {
		syntax, ok := CommandSyntax[ctx.Args()[0]]

		if ok {
			return ctx.Reply(syntax)
		} else {
			return ctx.Reply("Unknown command.")
		}
	}
}

// Syntax:
//
//	- /recall <ID/reply-to-message>
//	- /recall <username/name> <value>
func RecallHandler(ctx tele.Context) error {
	var (
		id    int64
		value string

		field string

		users []User

		parse_err error
		data_err  error

		length = len(ctx.Args())
	)

	switch length {
	case 0:
		if ctx.Message().ReplyTo == nil || ctx.Message().ReplyTo.Sender == nil {
			return ctx.Reply("ID required.")
		}

		id = ctx.Message().ReplyTo.Sender.ID
		field = "id"
	case 1:
		if id, parse_err = strconv.ParseInt(ctx.Args()[0], 0, 64); parse_err != nil {
			return ctx.Reply("Invalid ID.")
		}

		field = "id"
	case 2:
		field, value = ctx.Args()[0], ctx.Args()[1]
	default:
		field = ctx.Args()[0]

		if field == "name" {
			value = strings.Join(ctx.Args()[1:], " ")
		} else if field == "username" {
			value = ctx.Args()[1]
		}
	}

	switch field {
	case "id":
		var user User
		user, data_err = Data.FindByID(id)

		if data_err != nil {
			log.Printf("error finding user: %v\n", data_err)
		} else {
			users = []User{user}
		}
	case "name":
		users, data_err = Data.Filter(bson.D{{Key: "names", Value: value}})

		if data_err != nil {
			log.Printf("error querying users by name: %v\n", data_err)
		}
	case "username":
		users, data_err = Data.Filter(bson.D{{Key: "usernames", Value: strings.TrimLeft(value, "@")}})

		if data_err != nil {
			log.Printf("error querying users by name: %v\n", data_err)
		}
	default:
		return ctx.Reply("Unknown field name: \"" + field + "\"")
	}

	if len(users) == 0 {
		return ctx.Reply("No users were found.")
	} else if len(users) == 1 {
		d := DisplayUser(&users[0])

		if len(d) > 4096 {
			return ctx.Reply("The result's length exceeds the message size limit.", UploadResultBtnKeyboard)
		}

		return ctx.Reply(d, tele.ModeHTML)
	} else {
		str := make([]string, 0, len(users))

		for _, u := range users {
			str = append(
				str,
				fmt.Sprintf("[<code>%d</code>] %s", u.TelegramID, BoolToStr(len(u.Names) > 0, u.Names[len(u.Names)-1], "")),
			)
		}

		return ctx.Reply(
			fmt.Sprintf("<b>%d</b> users matched:\n\n\t- %s", len(users), strings.Join(str, "\n\t- ")),
			tele.ModeHTML,
		)
	}
}

func PermHelpBtnHandler(c tele.Context) error {
	perm, ok := CommandSyntax[CMD_PERM]

	if !ok {
		panic(errors.New("command \"" + CMD_PERM + "\" is not registered"))
	}

	return c.Edit(perm, BackToHelpKeyboard)
}

func BackToHelpBtnHandler(ctx tele.Context) error {
	return ctx.Edit(
		"Don't you hate it when poeple constantly change their names, usernames, and even their Telegram IDs, "+
			"and then you tend to forget who they were and what they did?\n"+
			"With Botone, you can keep track of their identities and record "+
			"their most significant actions, so you don't have to worry about forgetting and feeling like "+
			"everyone on telegram is the same person.\n\n"+
			"Click on the buttons below, to learn each command.",
		HelpMainPageKeyboard, tele.ModeHTML,
	)
}

func RegHelpBtnHandler(ctx tele.Context) error {
	return ctx.Edit("Register new users.\n\nSyntax:\n\n\t"+
		"- /reg <ID/reply-to-message>", BackToHelpKeyboard)
}

func CancelOperatorConfirmationBtnHandler(c tele.Context) error {
	return c.Edit("Operation cancelled.")
}

func ConfirmOperatorBtnHandler(c tele.Context) error {
	user_to_confirm, parse_err := strconv.ParseInt(c.Callback().Data, 0, 64)

	if parse_err != nil {
		log.Printf("error parsing ID: %v\n", parse_err)
		return c.Edit("Invalid callback data.")
	}

	user, data_err := Data.FindByID(user_to_confirm)

	if data_err != nil {
		log.Printf("error querying user ID: %v\n", data_err)
		return c.Edit("Could not perform this operation.")
	}

	user.Permission = 3
	err := Data.ReplaceByID(user_to_confirm, user)

	if err != nil {
		log.Printf("error updating user permission: %v\n", err)
		return c.Edit("Could not perform this action.")
	}

	return c.Edit("User is now an operator!")
}

func RecordHelpBtnHandler(ctx tele.Context) error {
	s, ok := CommandSyntax[CMD_RECORD]

	if !ok {
		panic(errors.New("command \"" + CMD_RECORD + "\" is not registered"))
	}

	return ctx.Edit(s, BackToHelpKeyboard)
}

func SetHelpBtnHandler(c tele.Context) error {
	s, ok := CommandSyntax[CMD_SET]

	if !ok {
		panic(errors.New("command \"" + CMD_SET + "\" is not registered"))
	}

	return c.Edit(s, BackToHelpKeyboard)
}

func AliasHelpBtnHandler(ctx tele.Context) error {
	s, ok := CommandSyntax[CMD_ALIAS]

	if !ok {
		panic(errors.New("command \"" + CMD_ALIAS + "\" is not registered"))
	}
	return ctx.Edit(s, BackToHelpKeyboard)
}

func RecallHelpBtnHandler(ctx tele.Context) error {
	s, ok := CommandSyntax[CMD_RECALL]

	if !ok {
		panic(errors.New("command \"" + CMD_RECALL + "\" is not registered"))
	}

	return ctx.Edit(s, BackToHelpKeyboard)
}

func UploadResultBtnHandler(ctx tele.Context) error {
	err := os.WriteFile(BUFF_FILE_PATH, []byte(StringBuffer), 0664)

	if err != nil {
		return ctx.Respond(&tele.CallbackResponse{Text: "An error has occurred."})
	}

	year, month, day := time.Now().Date()

	f := tele.Document{
		File:     tele.FromDisk(BUFF_FILE_PATH),
		FileName: fmt.Sprintf("Result-%4d-%s-%2d.txt", year, month.String(), day),
	}

	ctx.Delete()
	_, err = ctx.Bot().Send(ctx.Chat(), &f)
	return err
}

// Syntax:
// 	- /alias <ID/reply-to-message> <add/remove> <name/ID/username> <value1>; <value2> ..
//
// TODO:
//	Must implement server-side duplicate checking.
func AliasHandler(ctx tele.Context) error {
	var (
		length = len(ctx.Args())
		remove = false
		mode   = ""
		values = make([]string, 0, 1)

		id int64

		b_err error
		p_err error

		fetch = func(index int) {
			remove, b_err = StrToBool(ctx.Args()[index], "remove", "add")
			mode = ctx.Args()[index+1]
			values = strings.Split(strings.Join(ctx.Args()[index+2:], " "), ";")
		}
	)

	if length < 3 {
		return ctx.Reply("Insufficient arguments.")
	} else if length == 3 {
		// assuming that the id was given via message reply

		if ctx.Message().ReplyTo == nil || ctx.Message().ReplyTo.Sender == nil {
			return ctx.Reply("ID required.")
		}

		fetch(0)

		id = ctx.Message().ReplyTo.Sender.ID
	} else {
		// Acquire ID first

		if id, p_err = strconv.ParseInt(ctx.Args()[0], 0, 64); p_err == nil {
			fetch(1)
		} else if ctx.Message().ReplyTo != nil && ctx.Message().ReplyTo.Sender != nil {
			fetch(0)

			id = ctx.Message().ReplyTo.Sender.ID
		} else {
			return ctx.Reply("ID required.")
		}
	}

	// Check for errors

	if b_err != nil {
		log.Printf("error parsing string to boolean: %v\n", b_err)
		return ctx.Reply("Unknown add/remove value.")
	}

	if p_err != nil {
		return ctx.Reply("Invalid ID.")
	}

	user, u_err := Data.FindByID(id)

	if u_err != nil {
		log.Printf("error querying user by ID: %v\n", u_err)
		return ctx.Reply("ID not found.")
	}

	if mode == "name" {

		filtered_values := Undupe(values, user.Names)
		err := Data.Names(remove, id, filtered_values...)

		if err != nil {
			log.Printf(
				"error at %s%s %s request: %v\n",
				mode, BoolToStr(len(values) > 1, "s", ""),
				BoolToStr(remove, "pull", "push"),
				err,
			)

			return ctx.Reply("Could not perform this action.")
		}
	} else if mode == "username" {

		filtered_values := Undupe(values, user.Usernames)
		usernames := make([]string, 0, len(values))

		for _, v := range filtered_values {
			usernames = append(usernames, strings.TrimLeft(v, "@"))
		}

		err := Data.Usernames(remove, id, usernames...)

		if err != nil {
			log.Printf(
				"error at %s%s %s request: %v\n",
				mode, BoolToStr(len(values) > 1, "s", ""),
				BoolToStr(remove, "pull", "push"),
				err,
			)

			return ctx.Reply("Could not perform this action.")
		}
	} else if mode == "id" {

		filtered_values := Undupe(Map(values, func(s string) (int64, error) {
			i, e := strconv.ParseInt(s, 0, 64)

			if e == nil {
				return i, nil
			} else {
				log.Printf("error parsing string \"%s\": %v\n", s, e)
				return 0, e
			}
		}), user.AliasIDs)

		err := Data.Aliases(remove, id, filtered_values...)

		if err != nil {
			log.Printf("error sending %s request: %v\n", BoolToStr(remove, "pull", "push"), err)
			return ctx.Reply("Could not perform this action.")
		}
	} else {
		return ctx.Reply("Unknown mode value.")
	}

	return ctx.Reply(fmt.Sprintf("Alias%s %s.", BoolToStr(len(values) > 1, "es", ""), BoolToStr(remove, "removed", "added")))
}

// Syntax:
//
//	- /record <ID/reply-to-message> <category> [note1; note2; note3; ...]
func RecordHandler(ctx tele.Context) error {
	var (
		length = len(ctx.Args())

		parse_err error

		category string

		notes  []string
		record Record

		id int64
	)

	if length < 1 {
		return ctx.Reply("Insufficient arguments.")
	} else if length == 1 {
		if ctx.Message().ReplyTo == nil || ctx.Message().ReplyTo.Sender == nil {
			return ctx.Reply("ID required.")
		} else {
			id = ctx.Message().ReplyTo.Sender.ID
		}
	} else {
		if id, parse_err = strconv.ParseInt(ctx.Args()[0], 0, 64); parse_err != nil {
			log.Printf("error when parsing ID: %v\n", parse_err)
			return ctx.Reply("Invalid ID.")
		} else {
			category = ctx.Args()[1]

			joined := strings.Join(ctx.Args()[2:], " ")

			notes = strings.Split(joined, ";")

			for i, s := range notes {
				notes[i] = strings.Trim(s, " ")
			}
		}
	}

	record = Record{
		ChatID: ctx.Chat().ID,
		Notes:  notes,
		Date:   time.Now(),
	}

	f_user, f_err := Data.FindByID(id)

	if f_err != nil {
		return ctx.Reply("This user isn't registered.")
	}

	_, ok := f_user.Records[category]

	if ok {
		f_user.Records[category] = append(f_user.Records[category], record)
	} else {
		f_user.Records[category] = []Record{record}
	}

	err := Data.ReplaceByID(id, f_user)

	if err != nil {
		log.Printf("error replacing by ID: %v\n", err)
		return ctx.Reply("Could not complete this action.")
	}

	return ctx.Reply("Recorded.")
}

// Syntax:
//
//	- /rec <ID/reply-to-message>
func RegHandler(ctx tele.Context) error {
	var (
		id int64

		parse_err error
		data_err  error

		sender *tele.User
		user   User
	)

	if len(ctx.Args()) == 0 {
		if ctx.Message().ReplyTo != nil && ctx.Message().ReplyTo.Sender != nil {
			sender = ctx.Message().Sender
			id = ctx.Message().ReplyTo.Sender.ID
		} else {
			return ctx.Reply("ID required.")
		}
	} else {
		if id, parse_err = strconv.ParseInt(ctx.Args()[0], 0, 64); parse_err != nil {
			return ctx.Reply("Invalid ID.")
		}
	}

	_, data_err = Data.FindByID(id)

	if data_err == nil {
		return ctx.Reply("User is already registered.")
	}

	user = User{
		ID:         primitive.NewObjectID(),
		TelegramID: id,
		Names:      make([]string, 0, 1),
		Usernames:  make([]string, 0, 1),
		AliasIDs:   make([]int64, 0),
		Records:    map[string][]Record{},
	}

	if sender != nil {
		name, username, isBot := sender.FirstName+" "+sender.LastName, sender.Username, sender.IsBot

		if isBot {
			return ctx.Reply("The user is a bot; can't register bots.")
		}

		user.Names = append(user.Names, name)
		user.Usernames = append(user.Usernames, username)
	}

	data_err = Data.Add(user)

	if data_err != nil {
		log.Printf("error registering user: %v", data_err)
		return ctx.Reply("Could not perform this operation.")
	}

	return ctx.Reply("User registered.")
}

func UnregHandler(ctx tele.Context) error {
	var (
		id int64

		parse_err error
	)

	if len(ctx.Args()) == 0 {
		if ctx.Message().ReplyTo == nil || ctx.Message().ReplyTo.Sender == nil {
			return ctx.Reply("ID required.")
		}
		id = ctx.Message().ReplyTo.Sender.ID
	} else {
		if id, parse_err = strconv.ParseInt(ctx.Args()[0], 0, 64); parse_err != nil {
			return ctx.Reply("Invalid ID.")
		}
	}

	if Data.RemoveByID(id) != nil {
		log.Printf("error removing a user by ID: %v\n", id)
		return ctx.Reply("User not found.")
	} else {
		return ctx.Reply("User removed.")
	}
}

func QueryHandler(ctx tele.Context) error {
	var (
		str    string
		id     int64
		is_int bool

		users    []User
		data_err error
	)

	str, id, is_int = Parse(ctx.Query().Text)

	if is_int {
		user, data_err := Data.FindByID(id)
		fmt.Printf("%+v", user)

		if data_err == nil {
			users = []User{user}
		}

	} else {
		str = strings.TrimLeft(str, "@")

		users, data_err = Data.Filter(bson.D{{Key: "$or", Value: []bson.M{
			{"names": str},
			{"usernames": str},
		}}})
	}

	results := make(tele.Results, 0, len(users))

	if data_err != nil {
		fmt.Printf("error: %v\n", data_err)
		return ctx.Answer(nil)
	} else {
		for _, u := range users {
			name, id := "", u.TelegramID

			if len(u.Names) > 0 {
				name = u.Names[len(u.Names)-1]
			}

			results = append(results, &tele.ArticleResult{
				Title: BoolToStr(name != "", name, fmt.Sprintf("%d", id)),
				Text:  DisplayUser(&u),
			})
		}
	}

	for i := range results {
		results[i].SetResultID(strconv.Itoa(i))
		results[i].SetParseMode(tele.ModeHTML)
	}

	fmt.Printf("Results: %v", results)

	return ctx.Answer(&tele.QueryResponse{
		Results:   results,
		CacheTime: 60,
	})
}

func UnregHelpBtnHandler(ctx tele.Context) error {
	syntax, ok := CommandSyntax[CMD_UNREG]

	if !ok {
		panic(errors.New("command \"" + CMD_UNREG + "\" is not registered"))
	}

	return ctx.Edit(syntax, BackToHelpKeyboard)
}

// Allow/disallow a registered user to access the database.
//
// Syntax:
//
//	- /set <ID/reply-to-message> <permission-level>
//
// Permission:
//	- none / 0
//	- read / 1
//	- write / 2
func SetHandler(c tele.Context) error {
	var (
		id             int64
		permission     string
		permission_int int
		user           User

		parse_err      error
		permission_err error
		data_err       error
	)

	// Obtain ID

	if len(c.Args()) == 0 {
		return c.Reply("Insufficient arguments.")
	} else if len(c.Args()) == 1 {
		permission = c.Args()[0]

		if c.Message().ReplyTo == nil || c.Message().ReplyTo.Sender == nil {
			log.Printf("REPLYTO: %v\nSENDER: %v\n", c.Message().ReplyTo, c.Message().ReplyTo.Sender)
			return c.Reply("ID required.")
		} else {
			id = c.Message().ReplyTo.Sender.ID
		}
	} else {
		permission = c.Args()[1]
		id, parse_err = strconv.ParseInt(c.Args()[0], 0, 64)

		if parse_err != nil {
			return c.Reply("Invalid ID.")
		}
	}

	permission_int, permission_err = strconv.Atoi(permission)

	switch {
	case parse_err == nil &&
		(permission_err == nil ||
			permission == "none" ||
			permission == "read" ||
			permission == "write"):
	default:
		return c.Reply("Invalid permission value.")
	}

	// Query user

	user, data_err = Data.FindByID(id)

	if data_err != nil {
		log.Printf("error when querying a user by ID: %v\n", data_err)
		return c.Reply("User not found.")
	}

	if user.TelegramID == Config.OwnerTelegramID {
		return c.Reply("You're the owner; you can't revoke your own access.")
	}

	if permission_err != nil {
		switch permission {
		case "none":
			permission_int = 0
		case "read":
			permission_int = 1
		case "write":
			permission_int = 2
		}
	}

	if permission_int == 3 {
		OperatorConfirmationKeyboard.InlineKeyboard[0][1].Data = fmt.Sprintf("%d", id)

		return c.Reply(
			"You're about to grant this user <b>operator</b> access. Are you sure?",
			OperatorConfirmationKeyboard,
			tele.ModeHTML) ///////
	}

	user.Permission = permission_int

	err := Data.ReplaceByID(id, user)

	if err != nil {
		log.Printf("error updating user: %v\n", err)
		return c.Reply("Could not perform this action.")
	}

	return c.Reply("Permission set.")
}

func PermHandler(c tele.Context) error {
	var (
		id   int64
		perm string
		u    User

		parse_err error
	)

	if len(c.Args()) == 0 {
		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			id = c.Message().ReplyTo.Sender.ID
		} else {
			return c.Reply("ID required.")
		}
	} else {
		id, parse_err = strconv.ParseInt(c.Args()[0], 0, 64)

		if parse_err != nil {
			return c.Reply("Invalid ID.")
		}
	}

	u, _ = Data.FindByID(id)

	switch u.Permission {
	case 1:
		perm = "read-only"
	case 2:
		perm = "read/write"
	case 3:
		perm = "operator"
	default:
		perm = "no"
	}

	return c.Reply("This user has <b>"+perm+"</b> access.", tele.ModeHTML)
}
