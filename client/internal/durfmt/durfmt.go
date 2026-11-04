// Package durfmt formats durations for user-facing status output.
package durfmt

import (
	"fmt"
	"time"
)

// FormatRemainingDuration renders a time.Duration for the "Session expires"
// line. Examples: "2h 15m", "47m 12s", "8s", "expired 3m ago".
//
// Granularity drops to seconds only under a minute, otherwise minutes are
// the smallest unit shown — sub-minute precision is noise for a deadline
// that's hours or days out.
func FormatRemainingDuration(d time.Duration) string {
	if d <= 0 {
		return "expired " + HumaniseDuration(-d) + " ago"
	}
	return HumaniseDuration(d)
}

// HumaniseDuration renders a positive duration in compact form (e.g.
// "2h 15m", "47m", "8s"). Exposed alongside FormatRemainingDuration so
// callers that don't need the "expired … ago" wording can format
// positive durations directly.
func HumaniseDuration(d time.Duration) string {
	if d < time.Minute {
		s := int(d.Round(time.Second).Seconds())
		if s < 1 {
			s = 1
		}
		return fmt.Sprintf("%ds", s)
	}

	const (
		day    = 24 * time.Hour
		hour   = time.Hour
		minute = time.Minute
	)

	days := int64(d / day)
	d -= time.Duration(days) * day
	hours := int64(d / hour)
	d -= time.Duration(hours) * hour
	minutes := int64(d / minute)

	switch {
	case days > 0:
		if hours == 0 {
			return fmt.Sprintf("%dd", days)
		}
		return fmt.Sprintf("%dd %dh", days, hours)
	case hours > 0:
		if minutes == 0 {
			return fmt.Sprintf("%dh", hours)
		}
		return fmt.Sprintf("%dh %dm", hours, minutes)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}
