package utils

import (
	"fmt"
	"time"
)

// FormatDuration formats a given time duration into a string representation of hours, minutes, and seconds.
// It rounds the duration to the nearest second and uses the format "HH horas MM minutos y SS segundos".
// The function takes a time.Duration parameter 'd' and returns the formatted string.
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d horas %02d minutos y %02d segundos", h, m, s)
}
