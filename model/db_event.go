package model

import (
	"time"

	"github.com/google/uuid"
)

type DbEvent struct {
	Id         uuid.UUID
	Op_type    string
	Name       string
	Namespace  string
	Kind       string
	Message    string
	Reason     string
	Host       string
	Event      string
	First_time time.Time
	Last_time  time.Time
	Event_time time.Time
}
