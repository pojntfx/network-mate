// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2

package models

import (
	"time"
)

type JournalEntry struct {
	ID     int32
	Title  string
	Date   time.Time
	Body   string
	Rating int32
}
