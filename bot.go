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
	BTN_SET_PERM      = "setPermissionBtn"

	BTN_DELETE_ENTRY = "deleteEntryBtn"

	BTN_CANCEL_OPERATOR_CONFIRMATION = "cancelBtn"
	BTN_CONFIRM_OPERATOR             = "confirmOperatorBtn"

	// Help strings

	CREDITS = "<b>Botone v%s</b>\n\nCreator: <b>Henry Markle</b>\n" +
		"Powered by: <b>Go v1.18</b> & <b>MongoDB</b>\n\n" +
		"Code dependencies:\n\n\t" +
		"<b>Telebot</b> by <a href=\"https://github.com/tucnak\">tucnak</a>\n\t" +
		"<b>Mongo Go Driver</b> by <a href=\"https://www.mongodb.com/\">MongoDB</a>"

	HELP_MAIN = "Don't you hate it when poeple constantly change their names, usernames, and even their Telegram IDs, " +
		"and then you tend to forget who they were and what they did?\n" +
		"With Botone, you can keep track of their identities and record " +
		"their most significant actions, so you don't have to worry about forgetting and feeling like " +
		"everyone on telegram is the same person.\n\n" +
		"Click on the buttons below, to learn each command."

	HELP_RECALL = "Recall information about a person who's registered before. " +
		"You can use IDs, usernames, or names.\n\nSyntax:\n\n" +
		"- /recall <ID/reply-to-message>\n\n" +
		"- /recall name <name>\n\nExamples:\n\n" +
		"- /recall username <username>\n\n" +
		"/recall 69696969\n/recall name Miles Edgeworth"
	HELP_REG = "Register new users.\n\nSyntax:\n\n/reg <ID/reply-to-message> [description]"

	HELP_UNREG = "There are some people you just want to forget.\n" +
		"Unregister and delete them from the database.\n\nSyntax:\n\n" +
		"- /unreg <ID/reply-to-message>"

	HELP_HELP = "Learn each command's syntax by typing /help followed by the name of the command.\n\nSyntax:\n\n" +
		"- /help\n- /help <command>"

	HELP_RECORD = "Write down what the user did under a certain category.\n\n" +
		"Syntax:\n\n/record <ID/reply-to-message> <category> [note1; note2; note3; ..]\n\nExample:\n\n" +
		"/record 69696969 bans shared a pirated movie; he blamed me for eating his sandwish"

	HELP_SET = "Set description to a user record.\n" +
		"Syntax:\n\n/set <ID/reply-to-message> <description>\n\n" +
		"Example:\n\n/set 6969669 My brother-in-law.\n"

	HELP_PERM = "You can allow others to use your bot, but " +
		"you may not want to let them have complete access. Therefore, the bot comes with a permission system " +
		"that allows you to control how much access you want to give, by granting certain people certain permissions.\n\n" +
		"Permissions are represented as numbers. The higher the number, the more a user can access.\n\n" +
		"The available permissions are:\n\n" +
		"0 - No access\n" +
		"1 - Read-only\n" +
		"2 - Read/Write\n" +
		"3 - Operator\n" +
		"4 - Owner\n\n" +
		"Each command requires at minimum permission level as follows:\n\n" +
		"%s" +
		"\n\nBy default, every newly registered user has permission level 0, " +
		"which means that they can't interract with the bot at all.\n\n" +
		"You can increase the amount of control they have, with the /perm command.\n\n" +
		"By granting them permission level 1, you only allow them to use the /recall command and inline queries.\n\n" +
		"Permission level 2 unlocks the rest of the commands for the user, except for the /perm.\n\n" +
		"Permission level 3 is the operator eccess permission. Users with this permission can grant or revoke others' " +
		"permissions, but they obviously can't grant others permission level 3. Only the owner of the bot can do that.\n\n" +
		"Syntax:\n\n" +
		"- /perm <ID/reply-to-message>\n- /perm <ID/reply-to-message> set <permission-level>"

	HELP_ALIAS = "Add more IDs, names, or usernames that belong to the same person.\n\nSyntax:\n\n" +
		"/alias <ID/reply-to-message> <add/remove> <id/name/username> <value1>; <value2> ..\n\nExample:\n\n" +
		"/alias 69696969 add name Henry Markle; Steward; Rose Smith"

	// Messages

	MSG_NO_MATCH          = "No match"
	MSG_ID_REQUIRED       = "ID required"
	MSG_ID_NOT_FOUND      = "ID not found"
	MSG_INVALID_ID        = "Invalid ID"
	MSG_INSUFFICIENT_ARGS = "Insufficient arguments"
	MSG_COULD_NOT_PERFORM = "Could not perform this action"
	MSG_UNAUTHORIZED      = "You're unauthorized to perform this action"

	ERR_FMT_ADD    = "error adding ID: %v"
	ERR_FMT_QUERY  = "error finding ID: %v"
	ERR_FMT_DELETE = "error deleting ID: %v"
	ERR_FMT_UPDATE = "error updating ID: %v"
	ERR_FMT_PARSE  = "error parsing string: %v"

	// Formatted messages

	//

	BUFF_FILE_PATH = "./b.txt"

	VERSION = "0.59"
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
		CMD_RECALL:  1,
		CMD_HELP:    1,
		CMD_CREDITS: 1,
		CMD_SET:     2,
		CMD_ALIAS:   2,
		CMD_RECORD:  2,
		CMD_UNREG:   2,
		CMD_REG:     2,
		CMD_PERM:    3,

		BTN_UPLOAD_RESULT:                1,
		BTN_BACK_TO_HELP:                 1,
		BTN_RECALL_HELP:                  1,
		BTN_RECORD_HELP:                  2,
		BTN_ALIAS_HELP:                   2,
		BTN_REG_HELP:                     2,
		BTN_UNREG_HELP:                   2,
		BTN_SET_HELP:                     2,
		BTN_DELETE_ENTRY:                 2,
		BTN_PERM_HELP:                    3,
		BTN_SET_PERM:                     3,
		BTN_CANCEL_OPERATOR_CONFIRMATION: 4,
		BTN_CONFIRM_OPERATOR:             4,

		tele.OnQuery: 1,
	}

	CommandSyntax = map[string]string{
		CMD_REG: HELP_REG,

		CMD_RECORD: HELP_RECORD,

		CMD_RECALL: HELP_RECALL,

		CMD_ALIAS: CMD_ALIAS,

		CMD_HELP: CMD_HELP,

		CMD_UNREG: CMD_UNREG,
		CMD_SET:   HELP_SET,
		CMD_PERM: fmt.Sprintf(HELP_PERM, strings.Join(MaptoSlice(Permissions, func(k string, v int) (string, error) {
			if !strings.HasSuffix(k, "Btn") && k != "\aquery" {
				return fmt.Sprintf("/%s: %d", k, v), nil
			} else {
				return "", errors.New("must be a command")
			}
		}), "\n")),
	}

	StringBuffer = ""

	// Buttons

	DeleteEntryBtn = &tele.Btn{
		Unique: BTN_DELETE_ENTRY,
		Text:   "Delete Entry",
	}

	SetPermBtn = &tele.Btn{
		Unique: BTN_SET_PERM,
	}

	CancelOperatorConfirmationBtn = &tele.Btn{
		Unique: BTN_CANCEL_OPERATOR_CONFIRMATION,
		Text:   "Cancel",
	}

	ConfirmOperatorBtn = &tele.Btn{
		Unique: BTN_CONFIRM_OPERATOR,
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

	SetPermKeyboard = func(isOwner bool, user int64) *tele.ReplyMarkup {
		var markup *tele.ReplyMarkup

		var (
			none_btn = tele.Btn{
				Unique: BTN_SET_PERM,
				Text:   "None",
				Data:   fmt.Sprintf("0:%d", user),
			}.Inline()

			readonly_btn = tele.Btn{
				Unique: BTN_SET_PERM,
				Text:   "Read-only",
				Data:   fmt.Sprintf("1:%d", user),
			}.Inline()

			read_write_btn = tele.Btn{
				Unique: BTN_SET_PERM,
				Text:   "Read/Write",
				Data:   fmt.Sprintf("2:%d", user),
			}.Inline()
		)

		if isOwner {
			markup = &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					{
						*none_btn,
						*readonly_btn,
						*read_write_btn,
					},
					{
						*tele.Btn{
							Unique: BTN_SET_PERM,
							Text:   "Operator",
							Data:   fmt.Sprintf("3:%d", user),
						}.Inline(),
					},
				},
			}
		} else {
			markup = &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					{
						*none_btn,
						*readonly_btn,
						*read_write_btn,
					},
				},
			}
		}

		return markup
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
			CREDITS,
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
				HELP_MAIN,
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

	// Acquire id and/or field, depending on args length

	switch length {
	case 0:
		if ctx.Message().ReplyTo == nil || ctx.Message().ReplyTo.Sender == nil {
			return ctx.Reply(MSG_ID_REQUIRED)
		}

		id = ctx.Message().ReplyTo.Sender.ID
		field = "id"
	case 1:
		if id, parse_err = strconv.ParseInt(ctx.Args()[0], 0, 64); parse_err != nil {
			return ctx.Reply(MSG_INVALID_ID)
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

	// Query user(s), based on field

	switch field {
	case "id":
		var user User
		user, data_err = Data.FindByID(id)

		if data_err != nil {
			log.Printf(ERR_FMT_QUERY+"\n", data_err)
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
		// If there's no match
		return ctx.Reply(MSG_NO_MATCH)
	} else if len(users) == 1 {
		// If there's exactly one match
		d := DisplayUser(&users[0])

		if len(d) > 4096 {
			return ctx.Reply("The result's length exceeds the message size limit.", UploadResultBtnKeyboard)
		}

		deleteBtn := *DeleteEntryBtn
		deleteBtn.Data = fmt.Sprintf("%d", id)

		var keyboard *tele.ReplyMarkup

		sender, _ := Data.FindByID(ctx.Sender().ID)

		// If the sender has read/write permissions, and was in PM, and
		// the queried user wasn't the owner or an operator, a delete buttons shows up.
		if (sender.Permission >= 2) &&
			(ctx.Chat().ID == ctx.Sender().ID) &&
			(users[0].Permission < 3) &&
			(users[0].TelegramID != Config.OwnerTelegramID) {

			keyboard = &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					{*deleteBtn.Inline()},
				},
			}
		}

		return ctx.Reply(d, keyboard, tele.ModeHTML)
	} else {
		// If there's more than one match

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
		HELP_MAIN,
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

	// logging

	name := c.Message().Sender.FirstName + " " + c.Message().Sender.LastName

	ChanLogf("#op_confirm #perm\n[<code>%d</code>] %shas granted ID <code>%d</code> operator access.",
		Config.OwnerTelegramID,
		BoolToStr(name != "", name+" ", ""),
		user_to_confirm,
	)

	// returning

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

		if !remove {
			values = Undupe(values, user.Names)
		}

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

		if !remove {
			usernames = Undupe(usernames, user.Usernames)
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

		var filtered_values []int64
		mapped := Map(values, func(s string) (int64, error) {
			i, e := strconv.ParseInt(s, 0, 64)

			if e == nil {
				return i, nil
			} else {
				log.Printf("error parsing string \"%s\": %v\n", s, e)
				return 0, e
			}
		})

		if !remove {
			filtered_values = Undupe(mapped, user.AliasIDs)
		} else {
			filtered_values = mapped
		}

		err := Data.Aliases(remove, id, filtered_values...)

		if err != nil {
			log.Printf("error sending %s request: %v\n", BoolToStr(remove, "pull", "push"), err)
			return ctx.Reply("Could not perform this action.")
		}
	} else {
		return ctx.Reply("Unknown mode value.")
	}

	// logging

	name := ctx.Message().Sender.FirstName + " " + ctx.Message().Sender.LastName

	ChanLogf("#alias\n[<code>%d</code>] %shas %s alias%s %s ID <code>%d</code>:\n\t- %s",
		ctx.Message().Sender.ID,
		BoolToStr(name != "", name+" ", ""),
		BoolToStr(remove, "removed", "added"),
		BoolToStr(len(values) > 1, "es", ""),
		BoolToStr(remove, "from", "to"),
		id,
		strings.Join(values, "\n\t- "),
	)

	// returning

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

	// logging

	name := ctx.Message().Sender.FirstName + " " + ctx.Message().Sender.LastName

	ChanLogf("#record\n[<code>%d</code>] %shas recorded ID <code>%d</code>:\n\n%s",
		ctx.Message().Sender.ID,
		BoolToStr(name != "", name+" ", ""),
		id,
		RecordToStr(record, ""),
	)

	// returning

	return ctx.Reply("Recorded.")
}

// Syntax:
//
//	- /rec <ID/reply-to-message> [description]
func RegHandler(ctx tele.Context) error {
	var (
		id int64

		parse_err error
		data_err  error

		sender *tele.User
		user   User

		description = ""
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

		if len(ctx.Args()) >= 2 {
			description = strings.Join(ctx.Args()[1:], " ")
		}
	}

	_, data_err = Data.FindByID(id)

	if data_err == nil {
		return ctx.Reply("User is already registered.")
	}

	user = User{
		ID:          primitive.NewObjectID(),
		TelegramID:  id,
		Names:       make([]string, 0, 1),
		Usernames:   make([]string, 0, 1),
		AliasIDs:    make([]int64, 0),
		Description: description,
		Records:     map[string][]Record{},
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

	// logging

	name := ctx.Message().Sender.FirstName + " " + ctx.Message().Sender.LastName

	ChanLogf("#reg\n[<code>%d</code>] %shas registered ID <code>%d</code>.",
		ctx.Message().Sender.ID,
		BoolToStr(name != "", name+" ", ""),
		id,
	)

	// returning

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

	// You can't remove the owner

	if id == Config.OwnerTelegramID {
		return ctx.Reply("You can't remove the owner's ID.")
	}

	// You can't remove a user that's not registered

	u, ee := Data.FindByID(id)

	if ee != nil {
		log.Printf(ERR_FMT_QUERY+"\n", ee)
		return ctx.Reply(MSG_ID_NOT_FOUND)
	}

	// You can't remove a user with permission level 3, unless you're the owner

	if u.Permission >= 3 && (ctx.Sender().ID != Config.OwnerTelegramID) {
		return ctx.Reply("You need to be the owner, to remove an operator.")
	}

	if c, e := Data.RemoveByID(id); e != nil || c == 0 {
		log.Printf(ERR_FMT_DELETE+"\n", id)
		return ctx.Reply(MSG_ID_NOT_FOUND)
	} else {

		// logging

		name := ctx.Message().Sender.FirstName + " " + ctx.Message().Sender.LastName

		ChanLogf("#unreg\n[<code>%d</code>] %shas unregistered ID <code>%d</code>.",
			ctx.Message().Sender.ID,
			BoolToStr(name != "", name+" ", ""),
			id,
		)

		// returning

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

		if data_err == nil {
			users = []User{user}
		}

	} else {
		if strings.HasPrefix(str, "@") {
			str = strings.TrimLeft(str, "@")
			users, data_err = Data.Filter(bson.D{{Key: "usernames", Value: str}})
		} else {
			users, data_err = Data.Filter(bson.D{{Key: "names", Value: str}})
		}
	}

	results := make(tele.Results, 0, len(users))

	if data_err == nil {
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

// Set a description
//
// Syntax:
//
//	- /set <ID/reply-to-message> <description>
func SetHandler(c tele.Context) error {
	var (
		id   int64
		user User

		desc string

		parse_err error
		data_err  error
	)

	switch len(c.Args()) {
	case 0:
		return c.Reply("Insufficient arguments.")
	case 1:
		// /set <reply-to-message> <description>
		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			id = c.Message().ReplyTo.Sender.ID
		} else {
			return c.Reply("Insufficient arguments.")
		}

		desc = c.Args()[0]
	default:
		// /set <ID> <description>
		id, parse_err = strconv.ParseInt(c.Args()[0], 0, 64)

		ind := 1

		if parse_err != nil {
			if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
				id = c.Message().ReplyTo.Sender.ID
				ind = 0
			} else {
				return c.Reply("ID invalid or missing.")
			}
		}

		desc = strings.Join(c.Args()[ind:], " ")
	}

	user, data_err = Data.FindByID(id)

	if data_err != nil {
		log.Printf("error when querying user ID: %v\n", data_err)
		return c.Reply("User not found.")
	}

	user.Description = desc

	err := Data.ReplaceByID(id, user)

	if err != nil {
		log.Printf("error when replacing user: %v\n", err)
		return c.Reply("Could not perform this operation.")
	}

	// logging

	name := c.Message().Sender.FirstName + " " + c.Message().Sender.LastName

	ChanLogf("#description\n[<code>%d</code>] %shas updated the description for ID <code>%d</code>:\n\n\"%s\"",
		c.Message().Sender.ID,
		BoolToStr(name != "", name+" ", ""),
		id,
		desc,
	)

	// returning

	return c.Reply("Description set.")
}

// Syntax:
//
//	- /perm <ID/reply-to-message>
//	- /perm <ID/reply-to-message> set <permission-level>
func PermHandler(c tele.Context) error {
	var (
		id   int64
		perm string
		u    User

		isOwner bool

		set      = false
		new_perm int

		perm_parse_err error
		parse_err      error
		data_err       error
	)

	if c.Message().Sender.ID == Config.OwnerTelegramID {
		isOwner = true
	}

	// Acquire ID

	switch len(c.Args()) {
	case 0:
		// /perm <reply-to-message>
		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			id = c.Message().ReplyTo.Sender.ID
		} else {
			return c.Reply("ID required.")
		}

	case 1:
		// /perm <ID>
		id, parse_err = strconv.ParseInt(c.Args()[0], 0, 64)

	case 2:
		// /perm <reply-to-message> set <permission-level>
		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			id = c.Message().ReplyTo.Sender.ID
		} else {
			return c.Reply("ID required.")
		}

		if c.Args()[0] == "set" {
			set = true
			new_perm, perm_parse_err = strconv.Atoi(c.Args()[1])
		} else {
			return c.Reply("Invalid operation: \"" + c.Args()[0] + "\".")
		}

	default:
		// /perm <ID> set <permission-level>
		id, parse_err = strconv.ParseInt(c.Args()[0], 0, 64)

		if c.Args()[1] == "set" {
			set = true
			new_perm, perm_parse_err = strconv.Atoi(c.Args()[2])
		} else {
			return c.Reply("Invalid operation: \"" + c.Args()[1] + "\".")
		}
	}

	if parse_err != nil {
		return c.Reply("Invalid ID.")
	}

	if perm_parse_err != nil {
		return c.Reply("Invalid permission level.")
	}

	u, data_err = Data.FindByID(id)

	if set {
		if data_err != nil {
			log.Printf("error when querying a user by ID: %v\n", data_err)
			return c.Reply("User not found.")
		}

		if u.TelegramID == Config.OwnerTelegramID {
			return c.Reply("You're the owner; you can't change your own permission level.")
		}

		if new_perm >= 4 {
			return c.Reply("You can't grant <b>owner</b> access to other.")
		} else if new_perm >= 3 {
			if isOwner {
				keyboard := *OperatorConfirmationKeyboard
				keyboard.InlineKeyboard[0][1].Data = fmt.Sprintf("%d", id)

				return c.Reply(
					"You're about to grant this user <b>operator</b> access. Are you sure?",
					OperatorConfirmationKeyboard,
					tele.ModeHTML)
			} else {
				return c.Reply("You must be the owner to grant others <b>operator</b> access.")
			}
		} else {
			u.Permission = new_perm

			err := Data.ReplaceByID(id, u)

			if err != nil {
				log.Printf("error updating user: %v\n", err)
				return c.Reply("Could not perform this action.")
			}

			// logging

			name := c.Message().Sender.FirstName + " " + c.Message().Sender.LastName

			ChanLogf(
				"#permission #%s\n[<code>%d</code>] %shas updated the permission level of ID <code>%d</code> to <b>%d</b>.",
				BoolToStr(isOwner, "owner", "operator"),
				c.Message().Sender.ID,
				BoolToStr(name != "", name+" ", ""),
				id,
				new_perm,
			)

			// returning

			return c.Reply("Permission set.")
		}
	} else {

		switch u.Permission {
		case 1:
			perm = "read-only"
		case 2:
			perm = "read/write"
		case 3:
			perm = "operator"
		case 4:
			perm = "owner"
		default:
			perm = "no"
		}

		var keyboard *tele.ReplyMarkup
		var edit_prompt = ""

		if c.Chat().ID == c.Sender().ID && (id != Config.OwnerTelegramID) {
			keyboard = SetPermKeyboard(c.Sender().ID == Config.OwnerTelegramID, id)
			edit_prompt = "\n\nYou can edit the user's permission:"
		}

		return c.Reply("This user has <b>"+perm+"</b> access."+edit_prompt, keyboard, tele.ModeHTML)
	}
}

func SetPermBtnHandler(c tele.Context) error {
	var (
		id   int64
		user User
		perm int

		isOwner bool

		parse_err error
		data_err  error
	)

	if c.Callback().Sender.ID == Config.OwnerTelegramID {
		isOwner = true
	}

	p, i, ok := strings.Cut(c.Callback().Data, ":")

	if !ok {
		return c.Edit("Error: unknown callback data values \"" + c.Callback().Data + "\".")
	}

	id, parse_err = strconv.ParseInt(i, 0, 64)

	if parse_err != nil {
		log.Printf("error parsing IDs: %v\n", parse_err)
		return c.Edit("Invalid ID.")
	}

	switch p {
	case "0":
		perm = 0
	case "1":
		perm = 1
	case "2":
		perm = 2
	case "3":
		ok := Authorize(c.Callback().Sender.ID, BTN_CONFIRM_OPERATOR)

		if !ok {
			return c.Respond(&tele.CallbackResponse{Text: "Authorization failed."})
		}

		keyboard := *OperatorConfirmationKeyboard
		keyboard.InlineKeyboard[0][1].Data = fmt.Sprintf("%d", id)

		return c.Edit(
			"You're about to grant this user <b>operator</b> access. Are you sure?",
			&keyboard,
			tele.ModeHTML,
		)
	default:
		return c.Edit("Error: unknown callback data value: \"" + c.Callback().Data + "\".")
	}

	user, data_err = Data.FindByID(id)

	if data_err != nil {
		log.Printf("error querying ID: %v\n", data_err)
		return c.Edit("User not found.")
	}

	user.Permission = perm

	err := Data.ReplaceByID(id, user)

	if err != nil {
		log.Printf("error updating user permission: %v\n", err)
		return c.Edit("Could not perform this action.")
	} else {

		// logging

		name := c.Callback().Sender.FirstName + " " + c.Callback().Sender.LastName

		ChanLogf(
			"#permission #%s\n[<code>%d</code>] %shas updated the permission level of ID <code>%d</code> to <b>%d</b>.",
			BoolToStr(isOwner, "owner", "operator"),
			c.Callback().Sender.ID,
			BoolToStr(name != "", name+" ", ""),
			id,
			perm,
		)

		// returning

		return c.Edit("Permission updated.")
	}
}

func DeleteEntryBtnHandler(c tele.Context) error {
	var (
		id    int64
		count int64

		err       error
		parse_err error
	)

	id, parse_err = strconv.ParseInt(c.Callback().Data, 0, 64)

	if parse_err != nil {
		log.Printf("error parsing ID: %v\n", parse_err)
		return c.Edit("Could not perform this action: Invalid ID.")
	}

	if id == Config.OwnerTelegramID {
		return c.Edit("You can't remove the owner's registary.")
	}

	count, err = Data.RemoveByID(id)

	if err != nil {
		log.Printf("error removing ID: %v\n", err)
		return c.Edit("Could not perform this action: Database error.")
	}

	if count == 0 {
		return c.Edit("ID not found.")
	} else {

		// logging

		name := c.Message().Sender.FirstName + " " + c.Message().Sender.LastName

		ChanLogf("#unreg\n[<code>%d</code>] %shas unregistered the ID <code>%d</code>.",
			c.Message().Sender.ID,
			BoolToStr(name != "", name+" ", ""),
			id,
		)

		// returning

		return c.Edit("User unregistered.")
	}
}
