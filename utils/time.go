package utils

import "time"

// IsMiniForumOpen mengembalikan true jika waktu saat ini berada di antara 19:00–21:00 WIB.
func IsMiniForumOpen() bool {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.Local
	}

	now := time.Now().In(loc)

	start := time.Date(now.Year(), now.Month(), now.Day(), 19, 0, 0, 0, loc)
	end := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, loc)

	return now.After(start) && now.Before(end)
}

