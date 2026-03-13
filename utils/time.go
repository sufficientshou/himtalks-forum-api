package utils

import "time"

// IsMiniForumOpen mengembalikan true jika waktu saat ini berada di antara 19:00–21:00 WIB.
// Menggunakan offset manual UTC+7 supaya tidak tergantung tzdata di dalam container.
func IsMiniForumOpen() bool {
	// Ambil waktu UTC lalu geser ke WIB (UTC+7)
	nowUTC := time.Now().UTC()
	nowWIB := nowUTC.Add(7 * time.Hour)

	start := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 19, 0, 0, 0, time.UTC)
	end := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 21, 0, 0, 0, time.UTC)

	// Bandingkan dalam "ruang" WIB yang sudah digeser
	return nowWIB.After(start) && nowWIB.Before(end)
}


