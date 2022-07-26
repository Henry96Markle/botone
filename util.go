package main

import (
	"errors"
	"fmt"
	"strings"
)

func TrimUsername(s string) string { return strings.TrimLeft(s, "@") }
func StrToBool(s string, tr string, fl string) (b bool, err error) {
	switch s {
	case tr:
		return true, nil
	case fl:
		return false, nil
	default:
		return false, errors.New("string is not a boolean")
	}
}

func BoolToStr(b bool, str1 string, str2 string) string {
	if b {
		return str1
	} else {
		return str2
	}
}

func RecordToStr(r Record, offset string) string {
	return fmt.Sprintf(
		"Date: %v\n%sChat ID: <code>%d</code>%s",
		r.Date,
		offset,
		r.ChatID,
		BoolToStr(len(r.Notes) > 0, "\n"+offset+"Notes:\n"+offset+"- "+strings.Join(r.Notes, "\n"+offset+"- "), ""))
}

func RecordStrArr(offset string, records ...Record) []string {
	str := make([]string, 0, len(records))

	for _, r := range records {
		str = append(str, RecordToStr(r, offset))
	}

	return str
}

func IntToStrSlice(a ...int64) []string {
	str := make([]string, 0, len(a))

	for _, v := range a {
		str = append(str, fmt.Sprintf("%d", v))
	}

	return str
}

func DisplayUser(user *User) string {
	records := make([]string, 0, len(user.Records))

	var (
		name     = ""
		username = ""

		names     = ""
		usernames = ""
	)

	if len(user.Names) > 1 {
		names = strings.Join(user.Names[:len(user.Names)-1], "\n\t- ")
	}

	if len(user.Names) > 0 {
		name = user.Names[len(user.Names)-1]
	}

	if len(user.Usernames) > 1 {
		usernames = strings.Join(user.Usernames[:len(user.Usernames)-1],
			"</code>\n\t- <code>")
	}

	if len(user.Usernames) > 0 {
		username = user.Usernames[len(user.Usernames)-1]
	}

	for k, v := range user.Records {
		records = append(records, fmt.Sprintf("<b>%s</b>:\n\t%s", k, strings.Join(RecordStrArr("\t", v...), "\n\n\t")))
	}

	return fmt.Sprintf(
		"<b>Name:</b> %s\n<b>Username:</b> %s\n<b>ID:</b> <code>%d</code>%s%s%s%s",
		BoolToStr(len(user.Names) > 0, name, ""),
		BoolToStr(len(user.Usernames) > 0, username, ""),
		user.TelegramID,
		BoolToStr(
			len(user.Names) > 1, // The the last element in user.Names slice won't be displayed here.
			"\n\nAlso known by the "+BoolToStr(len(user.Names) > 2, "names", "name")+":\n\t- "+names,
			"",
		),
		BoolToStr(
			len(user.Usernames) > 1,
			"\n\nHeld the follwing usernames:\n\t- <code>"+usernames+"</code>", "",
		),
		BoolToStr(
			len(user.AliasIDs) > 0,
			"\n\nAlias IDs: \n\t- <code>"+
				strings.Join(IntToStrSlice(user.AliasIDs...),
					"</code>\n\t- <code>")+"</code>", "",
		),
		BoolToStr(
			len(user.Records) > 0,
			fmt.Sprintf(
				"\n\nRecords:\n\n\t%s",
				strings.Join(records, "\n\n\t"),
			), ""),
	)
}

// NOT finished yet.
func Parse(input string) (args []string, options map[string]string) {
	splt := strings.Split(input, " ")

	args = []string{}

	if strings.Contains(input, "where") {
		options = map[string]string{}
	}

	w_found := false
	w_index := -1

	for i, s := range splt {
		if s == "where" {
			w_found = true
			w_index = i
			break
		} else {
			args = append(args, s)
		}
	}

	if w_found {
		w_index++
	}

	return
}
