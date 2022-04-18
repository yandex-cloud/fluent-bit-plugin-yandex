package model

import (
	"time"

	"google.golang.org/protobuf/types/known/structpb"
)

type WriteRequest struct {
	Destination *Destination
	Resource    *Resource
	Entries     []*Entry
	Defaults    *Defaults
}

type Destination struct {
	LogGroupID string
	FolderID   string
}

type Resource struct {
	Type string
	ID   string
}

type Entry struct {
	Timestamp   time.Time
	Level       string
	Message     string
	JSONPayload *structpb.Struct
}

type Defaults struct {
	Level       string
	JSONPayload *structpb.Struct
}
