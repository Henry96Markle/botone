package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	tele "github.com/Henry96Markle/telebot"
)

const (
	DATABASE_NAME   = "telegram"
	COLLECTION_NAME = "user-records"

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
		"Permission level 2 unlocks the rest of the commands for the user, except for /perm.\n\n" +
		"Permission level 3 is the operator eccess permission. Operators can grant or revoke others' " +
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
	// main.go
	Config *Configuration

	Bot *tele.Bot

	Data *Database

	TermSig chan os.Signal

	Polling = false
	// bot.go

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

	PermissionNames = map[int]string{
		0: "None",
		1: "Read-only",
		2: "Read/Write",
		3: "Operator",
		4: "Owner",
	}

	CommandSyntax = map[string]string{
		CMD_REG: HELP_REG,

		CMD_RECORD: HELP_RECORD,

		CMD_RECALL: HELP_RECALL,

		CMD_ALIAS: HELP_ALIAS,

		CMD_HELP: HELP_HELP,

		CMD_UNREG: HELP_UNREG,
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
