// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package benchling

import (
	"context"

	"cloudeng.io/file/content"
	"cloudeng.io/webapi/apis/benchling/benchlingsdk"
	"cloudeng.io/webapi/operations"
)

// Document represents the structure of information within benchling
// in terms of an a single indexable document.
type Document struct {
	Entry    benchlingsdk.Entry   // An actual data entry.
	Folder   benchlingsdk.Folder  // The folder containing the entry.
	Project  benchlingsdk.Project // The project containing the folder.
	DayNotes string
	Parents  []string                     // The parent folders of the folder containing the entry.
	Users    map[string]benchlingsdk.User // All users referenced in the entry, keyed by their userid.
}

const (
	DocumentType = content.Type("benchling.com/document")
	EntryType    = content.Type("benchling.com/entry")
	ProjectType  = content.Type("benchling.com/project")
	FolderType   = content.Type("benchling.com/folder")
	UserType     = content.Type("benchling.com/user")
)

type Entries struct {
	NextToken *string
	Entries   []benchlingsdk.Entry
}

type Users struct {
	NextToken *string
	Users     []benchlingsdk.User
}

type Folders struct {
	NextToken *string
	Folders   []benchlingsdk.Folder
}

type Projects struct {
	NextToken *string
	Projects  []benchlingsdk.Project
}

type Scanners interface {
	Entries | Users | Folders | Projects
}

type Params interface {
	*benchlingsdk.ListEntriesParams | *benchlingsdk.ListUsersParams | *benchlingsdk.ListFoldersParams | *benchlingsdk.ListProjectsParams
}

func NewScanner[ScannerT Scanners, ParamsT Params](ctx context.Context, serviceURL string, params ParamsT, opts ...operations.Option) *operations.Scanner[ScannerT] {
	pg := newPaginator[ScannerT](ctx, serviceURL, params)
	return operations.NewScanner(pg, opts...)
}

type Objects interface {
	benchlingsdk.Entry | benchlingsdk.User | benchlingsdk.Folder | benchlingsdk.Project | Document
}

func ObjectID[ObjectT Objects](obj ObjectT) string {
	switch c := (any)(obj).(type) {
	case benchlingsdk.Entry:
		return "entry:" + *c.Id
	case benchlingsdk.User:
		return "user:" + *c.Id
	case benchlingsdk.Folder:
		return "folder:" + *c.Id
	case benchlingsdk.Project:
		return "projec:" + *c.Id
	case Document:
		return "document:" + *c.Entry.Id
	}
	return ""
}

func ContentType[ObjectT Objects](obj ObjectT) content.Type {
	switch (any)(obj).(type) {
	case benchlingsdk.Entry:
		return EntryType
	case benchlingsdk.User:
		return UserType
	case benchlingsdk.Folder:
		return FolderType
	case benchlingsdk.Project:
		return ProjectType
	case Document:
		return DocumentType
	}
	return ""
}
