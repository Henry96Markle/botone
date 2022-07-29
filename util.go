package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Trims the "@" from the username string.
func TrimUsername(s string) string { return strings.TrimLeft(s, "@") }

// Parses a string value; if the string is equal to tr, true is returned, else if it matches lf,
// false is returned; otherwise an error is returned.
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

// Parses a User.Record into a formatted string.
func RecordToStr(r Record, offset string) string {
	return fmt.Sprintf(
		"Date: %v\n%sChat ID: <code>%d</code>%s",
		r.Date,
		offset,
		r.ChatID,
		BoolToStr(len(r.Notes) > 0, "\n"+offset+"Notes:\n"+offset+"- "+strings.Join(r.Notes, "\n"+offset+"- "), ""))
}

// Applies RecordToStr(Record, string) to a slice.
func RecordStrArr(offset string, records ...Record) []string {
	str := make([]string, 0, len(records))

	for _, r := range records {
		str = append(str, RecordToStr(r, offset))
	}

	return str
}

// Parses an slice of 64-bit integers to strings.
func IntToStrSlice(a ...int64) []string {
	str := make([]string, 0, len(a))

	for _, v := range a {
		str = append(str, fmt.Sprintf("%d", v))
	}

	return str
}

// Parses a user to a formatted string.
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
		"<b>Name:</b> %s\n<b>Username:</b> %s\n<b>ID:</b> <code>%d</code>%s%s%s%s%s",
		BoolToStr(len(user.Names) > 0, name, ""),
		BoolToStr(len(user.Usernames) > 0, username, ""),
		user.TelegramID,
		BoolToStr(len(user.Description) > 0, "\n\n"+user.Description, ""),
		BoolToStr(
			len(user.Names) > 1, // The the last element in user.Names slice won't be displayed here.
			"\n\nAlso known by the "+BoolToStr(len(user.Names) > 2, "names", "name")+":\n\t- "+names,
			"",
		),
		BoolToStr(
			len(user.Usernames) > 1,
			"\n\nHeld the follwing username"+BoolToStr(len(user.Usernames) > 2, "s", "")+":\n\t- <code>"+usernames+"</code>", "",
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

// Filter duplicate items from arr1, if they exist in arr2, and returns the a new array with the
// filtered elements.
func Undupe[K comparable](arr1, arr2 []K) []K {
	filtered := make([]K, 0, len(arr1))

outter:
	for _, x := range arr1 {
		for _, y := range arr2 {
			if x == y {
				continue outter
			}
		}
		filtered = append(filtered, x)
	}

	return filtered
}

// Applies a function to each element of an array and returns a new array with the resulted elements.
//The function may return an error instead
// of a new value. In that case, the element is skipped and won't be added to the result array.
func Map[A any, B any](arr []A, operator func(A) (B, error)) []B {
	result := make([]B, 0, len(arr))

	for _, e := range arr {
		r, err := operator(e)
		if err == nil {
			result = append(result, r)
		}
	}

	return result
}

// Tries to parse a string. If successful, a 64-bit integer is returned, otherwise the string is returned.
// If the parsing was successful, a boolean value of true would be returned, otherwise false.
func Parse(a string) (string, int64, bool) {
	id, p_err := strconv.ParseInt(a, 0, 64)

	if p_err != nil {
		return a, 0, false
	} else {
		return "", id, true
	}
}

// Turns a map into a slice
func MaptoSlice[A comparable, B any, K any](m map[A]B, operator func(A, B) (K, error)) []K {
	res := make([]K, 0, len(m))

	for k, v := range m {
		value, err := operator(k, v)

		if err == nil {
			res = append(res, value)
		}
	}

	return res
}

func Authorize(id int64, action string) bool {
	perm, ok := Permissions[action]

	if !ok {
		return false
	}

	u, err := Data.FindByID(id)

	if err != nil {
		return false
	}

	return u.Permission >= perm
}
