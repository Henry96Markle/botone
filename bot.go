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
	return ctx.Edit(HELP_REG, BackToHelpKeyboard)
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
			to_check := user.AliasIDs
			to_check = append(to_check, user.TelegramID)
			filtered_values = Undupe(mapped, to_check)
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
		return ctx.Reply(MSG_ID_NOT_FOUND)
	}

	if (id == Config.OwnerTelegramID || f_user.Permission >= 4) && (ctx.Message().Sender.ID != Config.OwnerTelegramID) {
		return ctx.Reply("You can't record an owner.")
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

	// Determine whether the strings is an ID, a name, or a username

	str, id, is_int = Parse(ctx.Query().Text)

	if is_int {
		user, data_err := Data.FindByID(id)
		more_users, data_err2 := Data.Filter(bson.D{{Key: "alias_ids", Value: id}})

		if data_err == nil || data_err2 == nil {
			users = make([]User, 0, len(users)+len(more_users))
		}

		if data_err == nil {
			users = append(users, user)
		}

		if data_err2 == nil {
			for _, m := range more_users {
				m.Description = "This user has the ID as an alias."
				users = append(users, m)
			}
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
				Title:       BoolToStr(name != "", name, fmt.Sprintf("%d", id)),
				Text:        DisplayUser(&u),
				Description: u.Description,
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
		return c.Reply(MSG_INSUFFICIENT_ARGS)
	case 1:
		// /set <reply-to-message> <description>
		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			id = c.Message().ReplyTo.Sender.ID
		} else {
			return c.Reply(MSG_INSUFFICIENT_ARGS)
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
		log.Printf(ERR_FMT_QUERY+"\n", data_err)
		return c.Reply(MSG_ID_NOT_FOUND)
	}

	if (id == Config.OwnerTelegramID || user.Permission >= 4) && c.Message().Sender.ID != Config.OwnerTelegramID {
		return c.Reply("You can't change owner's data.")
	}

	user.Description = desc

	err := Data.ReplaceByID(id, user)

	if err != nil {
		log.Printf(ERR_FMT_UPDATE+"\n", err)
		return c.Reply(MSG_COULD_NOT_PERFORM)
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
			return c.Reply(MSG_ID_REQUIRED)
		}

	case 1:
		// /perm <ID>
		id, parse_err = strconv.ParseInt(c.Args()[0], 0, 64)

	case 2:
		// /perm <reply-to-message> set <permission-level>
		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			id = c.Message().ReplyTo.Sender.ID
		} else {
			return c.Reply(MSG_ID_REQUIRED)
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
		return c.Reply(MSG_INVALID_ID)
	}

	if perm_parse_err != nil {
		return c.Reply("Invalid permission level.")
	}

	u, data_err = Data.FindByID(id)

	if set {
		if data_err != nil {
			log.Printf(ERR_FMT_QUERY+"\n", data_err)
			return c.Reply(MSG_ID_NOT_FOUND)
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
				log.Printf(ERR_FMT_UPDATE+"\n", err)
				return c.Reply(MSG_COULD_NOT_PERFORM)
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
		log.Printf(ERR_FMT_QUERY+"\n", data_err)
		return c.Edit(MSG_ID_NOT_FOUND)
	}

	user.Permission = perm

	err := Data.ReplaceByID(id, user)

	if err != nil {
		log.Printf("error updating user permission: %v\n", err)
		return c.Edit(MSG_COULD_NOT_PERFORM)
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
		log.Printf(ERR_FMT_PARSE+"\n", parse_err)
		return c.Edit(MSG_INVALID_ID)
	}

	if id == Config.OwnerTelegramID {
		return c.Edit("You can't remove the owner's registary.")
	}

	count, err = Data.RemoveByID(id)

	if err != nil {
		log.Printf(ERR_FMT_DELETE+"\n", err)
		return c.Edit("Could not perform this action: Database error.")
	}

	if count == 0 {
		return c.Edit(MSG_ID_NOT_FOUND)
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

func DelrecHelpBtnHandler(c tele.Context) error {
	s, ok := CommandSyntax[CMD_DELREC]

	if !ok {
		panic("command \"" + CMD_DELREC + "\" is not registered")
	}

	return c.Edit(s, BackToHelpKeyboard)
}

// Syntax:
//
//	/delrec <ID/reply-to-message> <category> [note-index] ..
func DelrecHandler(c tele.Context) error {
	var (
		id   int64
		user User

		category  = ""
		index     = ""
		index_int = -1

		all_recs_to_delete_exists = false
		user_display              string

		cat_to_delete        []Record
		cat_to_delete_exists = false

		rec_to_delete        Record
		rec_to_delete_exists = false

		parse_err  error
		parse_err2 error
		data_err   error
	)

	// Acquire ID & category & nore-index

	switch len(c.Args()) {
	// /delrec <reply-to-message>	-> delete all records of a user
	case 0:
		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			id = c.Message().ReplyTo.Sender.ID
			goto SkipValidation
		} else {
			return c.Reply(MSG_ID_REQUIRED)
		}
	// /delrec <ID> 							-> delete all records from a user
	// /delrec <reply-to-message> <category>	-> delete all records of a category from a user
	case 1:
		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			id = c.Message().ReplyTo.Sender.ID
			category = c.Args()[0]
		} else if id, parse_err = strconv.ParseInt(c.Args()[0], 0, 64); parse_err != nil {
			return c.Reply(MSG_ID_REQUIRED)
		}

	// /delrec <ID> <category>								-> delete all records from a user
	// /delrec <reply-to-message> <category> [note-index]	-> delete a single record from a category from a user
	case 2:
		if id, parse_err = strconv.ParseInt(c.Args()[0], 0, 64); parse_err == nil {
			category = c.Args()[1]
		} else if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			id = c.Message().ReplyTo.Sender.ID
			category = c.Args()[0]
			index = c.Args()[1]
		}

	// /delrec <ID> <category> [note-index]	-> delete a single record from a category from a user
	default:
		if id, parse_err = strconv.ParseInt(c.Args()[0], 0, 64); parse_err == nil {
			category = c.Args()[1]
			index = c.Args()[2]
		} else if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			id = c.Message().ReplyTo.Sender.ID
			category = c.Args()[0]
			index = c.Args()[1]
		}
	}

	// Validate index

	if index_int, parse_err2 = strconv.Atoi(index); index != "" && parse_err2 != nil {
		return c.Reply("Invalid index value.")
	} else {
		// Turn the index into zero-based
		index_int--
	}

	// Check if user exists

	user, data_err = Data.FindByID(id)

	if data_err != nil {
		log.Printf(ERR_FMT_QUERY+"\n", data_err)
		return c.Reply(MSG_ID_NOT_FOUND)
	}

	if (id == Config.OwnerTelegramID || user.Permission >= 4) && c.Message().Sender.ID != Config.OwnerTelegramID {
		return c.Reply("You can't modify owner's records.")
	}

	// Validate category if given

	if category != "" {
		k, exists := user.Records[category]

		if !exists {
			return c.Reply("Category \"" + category + "\" does not exist.")
		}

		// Check if note-index is within range, if given

		if index != "" && (index_int >= len(k) || index_int < 0) {
			return c.Reply("Note index is out of bounds.")
		}
	}

	// Start deleting

SkipValidation:

	switch len(c.Args()) {
	case 0:
		all_recs_to_delete_exists = true
		user_display = DisplayUser(&user)

		user.Records = map[string][]Record{}
	case 1:
		if category != "" {
			cat_to_delete_exists = true
			cat_to_delete = user.Records[category]

			delete(user.Records, category)
		} else {
			all_recs_to_delete_exists = true
			user_display = DisplayUser(&user)

			user.Records = map[string][]Record{}
		}
	case 2:
		if index != "" {
			rec_to_delete_exists = true
			rec_to_delete = user.Records[category][index_int]

			new_records := make([]Record, 0, len(user.Records[category])-1)
			for i, r := range user.Records[category] {
				if i != index_int {
					new_records = append(new_records, r)
				}
			}
			user.Records[category] = new_records
		} else {
			cat_to_delete_exists = true
			cat_to_delete = user.Records[category]

			delete(user.Records, category)
		}
	default:
		rec_to_delete_exists = true
		rec_to_delete = user.Records[category][index_int]

		new_records := make([]Record, 0, len(user.Records[category])-1)
		for i, r := range user.Records[category] {
			if i != index_int {
				new_records = append(new_records, r)
			}
		}
		user.Records[category] = new_records
	}

	// Send to database

	err := Data.ReplaceByID(id, user)
	if err != nil {
		log.Printf(ERR_FMT_UPDATE+"\n", err)
		return c.Reply(MSG_COULD_NOT_PERFORM)
	}

	// logging

	name := c.Message().Sender.FirstName + " " + c.Message().Sender.LastName

	ChanLogf("#delrec\n[<code>%d</code>] %shas deleted %s from ID %d.%s.",
		c.Sender().ID,
		BoolToStr(name != "", name+" ", ""),
		BoolToStr(
			index != "",
			"a record from category \""+category+"\"",
			BoolToStr(category != "", "a category", "all records"),
		),
		id,
		BoolToStr(
			index != "" && rec_to_delete_exists,
			"\n\nDeleted record:\n\n"+RecordToStr(rec_to_delete, ""),
			BoolToStr(
				category != "" && cat_to_delete_exists,
				"\n\nDeleted record category \""+category+"\":\n\n\t"+strings.Join(RecordStrArr("\t", cat_to_delete...), "\n"),
				BoolToStr(all_recs_to_delete_exists, "\n\n"+user_display, ""),
			),
		),
	)

	// returning

	return c.Reply(fmt.Sprintf(
		"%s removed.",
		BoolToStr(
			category != "",
			BoolToStr(index != "", "Record", "\""+category+"\" category"),
			"All records",
		),
	))
}
