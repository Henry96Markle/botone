package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	tele "gopkg.in/telebot.v3"
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

	// Button unique strings

	BTN_RECORD_HELP   = "recordCommandHelpBtn"
	BTN_REG_HELP      = "regCommandHelpBtn"
	BTN_ALIAS_HELP    = "aliasCommandHelpBtn"
	BTN_RECALL_HELP   = "recallCommandHelpBtn"
	BTN_UNREG_HELP    = "unregCommandHelpBtn"
	BTN_BACK_TO_HELP  = "backToHelpMainPageBtn"
	BTN_UPLOAD_RESULT = "uploadResultBtn"
	BTN_SET_HELP      = "setCommandHelpBtn"

	BUFF_FILE_PATH = "./b.txt"

	VERSION = "0.0.1"
)

var (
	Commands = []string{
		CMD_HELP, CMD_REG, CMD_RECORD, CMD_ALIAS, CMD_RECALL, CMD_UNREG, CMD_SET, CMD_CREDITS,
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

		BTN_RECORD_HELP:   2,
		BTN_ALIAS_HELP:    2,
		BTN_BACK_TO_HELP:  1,
		BTN_RECALL_HELP:   1,
		BTN_REG_HELP:      2,
		BTN_UNREG_HELP:    2,
		BTN_UPLOAD_RESULT: 1,
		BTN_SET_HELP:      3,
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
	}

	StringBuffer = ""

	// Buttons

	SetHelpBtn = &tele.Btn{
		Unique: BTN_SET_HELP,
		Text:   CMD_SET,
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

	UploadResultBtnKeyboard = &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{
			{*UploadResultBtn.Inline()},
		},
	}

	HelpMainPageKeyboard = &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{
			{*RecordHelpBtn.Inline(), *AliasHelpBtn.Inline(), *RecallHelpBtn.Inline()},
			{*RegHelpBtn.Inline(), *UnregHelpBtn.Inline(), *SetHelpBtn.Inline()},
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

func BackToHelpBtnHandler(ctx tele.Context) error {
	return ctx.Edit(
		"Don't you hate it when poeple constantly change their names, usernames, and even their Telegram IDs, "+
			"and then you tend to forget who they were and what they did?\n"+
			"With Botone, you can keep track of their identidies and record "+
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

	if mode == "name" {
		err := Data.Names(remove, id, values...)

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
		usernames := make([]string, 0, len(values))

		for _, v := range values {
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
		ids := make([]int64, 0, len(values))

		for _, v := range values {
			i, e := strconv.ParseInt(v, 0, 64)

			if e == nil {
				ids = append(ids, i)
			} else {
				log.Printf("error parsing string \"%s\": %v\n", v, e)
			}
		}

		err := Data.Aliases(remove, id, ids...)

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
		return ctx.Reply("User already exists in the database.")
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
	urls := []string{
		"http://photo.jpg",
		"http://photo2.jpg",
	}

	results := make(tele.Results, len(urls)) // []tele.Result
	for i, url := range urls {
		result := &tele.PhotoResult{
			URL:      url,
			ThumbURL: url, // required for photos
		}

		results[i] = result
		// needed to set a unique string ID for each result
		results[i].SetResultID(strconv.Itoa(i))
	}

	return ctx.Answer(&tele.QueryResponse{
		Results:   results,
		CacheTime: 60, // a minute
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

	user.Permission = permission_int

	err := Data.ReplaceByID(id, user)

	if err != nil {
		log.Printf("error updating user: %v\n", err)
		return c.Reply("Could not perform this aaction.")
	}

	return c.Reply("Permission set.")
}
