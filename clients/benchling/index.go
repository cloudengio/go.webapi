// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package benchling

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"cloudeng.io/file/content"
	"cloudeng.io/file/content/stores"
	"cloudeng.io/file/filewalk"
	"cloudeng.io/path"
	"cloudeng.io/webapi/clients/benchling/benchlingsdk"
	"cloudeng.io/webapi/operations"
)

type DocumentIndexer struct {
	fs          operations.FS
	downloads   string
	concurrency int
	sharder     path.Sharder
	mu          sync.Mutex // Locks users, projects, folders, entries
	users       map[string]benchlingsdk.User
	projects    map[string]benchlingsdk.Project
	folders     map[string]benchlingsdk.Folder
	entries     map[string]benchlingsdk.Entry
}

func NewDocumentIndexer(fs operations.FS, downloads string, sharder path.Sharder, concurrency int) *DocumentIndexer {
	return &DocumentIndexer{
		fs:          fs,
		downloads:   downloads,
		concurrency: concurrency,
		users:       make(map[string]benchlingsdk.User),
		projects:    make(map[string]benchlingsdk.Project),
		folders:     make(map[string]benchlingsdk.Folder),
		entries:     make(map[string]benchlingsdk.Entry),
		sharder:     sharder,
	}
}

func (di *DocumentIndexer) Index(ctx context.Context) error {
	return di.index(ctx)
}

func (di *DocumentIndexer) readFile(_ context.Context, prefix, name string, ctype content.Type, buf []byte, err error) error {
	if err != nil {
		log.Printf("benchling indexder: read %v: error: %v\n", di.fs.Join(prefix, name), err)
		return err
	}
	switch ctype {
	case EntryType:
		var obj content.Object[benchlingsdk.Entry, operations.Response]
		if err := obj.Decode(buf); err != nil {
			return err
		}
		di.mu.Lock()
		di.entries[ObjectID(obj.Value)] = obj.Value
		di.mu.Unlock()
	case ProjectType:
		var obj content.Object[benchlingsdk.Project, operations.Response]
		if err := obj.Decode(buf); err != nil {
			return err
		}
		di.mu.Lock()
		di.projects[ObjectID(obj.Value)] = obj.Value
		di.mu.Unlock()
	case FolderType:
		var obj content.Object[benchlingsdk.Folder, operations.Response]
		if err := obj.Decode(buf); err != nil {
			return err
		}
		di.mu.Lock()
		di.folders[ObjectID(obj.Value)] = obj.Value
		di.mu.Unlock()
	case UserType:
		var obj content.Object[benchlingsdk.User, operations.Response]
		if err := obj.Decode(buf); err != nil {
			return err
		}
		di.mu.Lock()
		di.users[ObjectID(obj.Value)] = obj.Value
		di.mu.Unlock()
	}
	return nil
}

func (di *DocumentIndexer) populate(ctx context.Context, prefix string, contents []filewalk.Entry, err error) error {
	if err != nil {
		if di.fs.IsNotExist(err) {
			return nil
		}
		return err
	}
	start := time.Now()
	store := stores.New(di.fs, di.concurrency)
	defer func() {
		nUsers := len(di.users)
		nEntries := len(di.entries)
		nFolders := len(di.folders)
		nProjects := len(di.projects)
		total := nUsers + nEntries + nFolders + nProjects
		log.Printf("benchling indexder: %v total read: %v (users: %v, entries %v, folders %v, projects %v): read %v", prefix, total, nUsers, nEntries, nFolders, nProjects, time.Since(start))
	}()

	names := make([]string, len(contents))
	for i, c := range contents {
		names[i] = c.Name
	}

	return store.ReadV(ctx, prefix, names, di.readFile)
}

func handleSimpleNotePart(note benchlingsdk.EntryNotePart) (string, bool, error) {
	switch benchlingsdk.SimpleNotePartType(note.Type) {
	case benchlingsdk.SimpleNotePartTypeCode,
		benchlingsdk.SimpleNotePartTypeListBullet,
		benchlingsdk.SimpleNotePartTypeListNumber:
		return "", true, nil
	case benchlingsdk.SimpleNotePartTypeText:
		sn, err := note.AsSimpleNotePart()
		if err != nil {
			return "", true, err
		}
		return *sn.Text, true, nil
	}
	return "", false, nil
}

func (di *DocumentIndexer) dayText(entry *benchlingsdk.Entry) string {
	var notes strings.Builder
	if entry.Days != nil {
		for _, dayEntry := range *entry.Days {
			for _, n := range *dayEntry.Notes {
				sn, handled, err := handleSimpleNotePart(n)
				if err != nil {
					log.Printf("benchling indexer: EntryDay_Notes_Item: %#v: %v", n, err)
					continue
				}
				if handled {
					notes.WriteString(sn)
					notes.WriteRune('\n')
					continue
				}
				note, err := n.ValueByDiscriminator()
				if err != nil {
					log.Printf("benchling indexer: EntryDay_Notes_Item: %#v: %v", n, err)
					continue
				}
				switch v := note.(type) {
				case *benchlingsdk.SimpleNotePart:
					notes.WriteString(*v.Text)
					notes.WriteRune('\n')
				case *benchlingsdk.TableNotePart:
					notes.WriteString(*v.Text)
					notes.WriteRune('\n')
				case *benchlingsdk.CheckboxNotePart:
					notes.WriteString(*v.Text)
					notes.WriteRune('\n')
				case *benchlingsdk.ExternalFileNotePart:
					notes.WriteString(*v.Text)
					notes.WriteRune('\n')
				}
			}
		}
	}
	return notes.String()
}

func (di *DocumentIndexer) index(ctx context.Context) error {
	err := filewalk.ContentsOnly(ctx, di.fs, di.downloads, di.populate)
	if err != nil {
		return err
	}
	join := di.fs.Join
	store := stores.New(di.fs, di.concurrency)
	defer store.Finish(ctx) //nolint:errcheck
	log.Printf("benchling indexer: %v entries\n", len(di.entries))
	n := 0
	last := time.Now()
	for _, entry := range di.entries {
		doc := Document{
			Entry:   entry,
			Folder:  di.folders["folder:"+*entry.FolderId],
			Project: di.projects["folder:"+*entry.FolderId],
			Users:   make(map[string]benchlingsdk.User),
		}
		doc.DayNotes = di.dayText(&entry)
		doc.Parents = di.parents(*entry.FolderId, nil)
		di.addUsers(&doc, entry.Authors)
		di.addUsers(&doc, entry.AssignedReviewers)
		if entry.Creator != nil {
			v := di.users["user:"+*entry.Creator.Id]
			doc.Users[*entry.Creator.Id] = v
			if v.Name == nil {
				log.Printf("failed to find user: %v\n", *entry.Creator.Id)
			}
		}
		obj := content.Object[Document, struct{}]{
			Type:     DocumentType,
			Value:    doc,
			Response: struct{}{},
		}
		id := ObjectID(doc)
		prefix, suffix := di.sharder.Assign(fmt.Sprintf("%v", id))
		prefix = join(di.downloads, prefix)
		if err := obj.Store(ctx, store, prefix, suffix, content.JSONObjectEncoding, content.GOBObjectEncoding); err != nil {
			log.Printf("benchling indexer: failed to write user: %v as %v %v: %v\n", id, prefix, suffix, err)
		}
		n++
		if n%100 == 0 {
			log.Printf("benchling indexer: written %v/%v: %v\n", n, len(di.entries), time.Since(last))
			last = time.Now()
		}
	}
	log.Printf("benchling indexer: written %v/%v: %v\n", n, len(di.entries), time.Since(last))
	return store.Finish(ctx)
}

func (di *DocumentIndexer) parents(id string, p []string) []string {
	n, ok := di.folders["folder:"+id]
	if !ok {
		return p
	}
	if n.ParentFolderId == nil {
		return append(p, *n.Name)
	}
	return di.parents(*n.ParentFolderId, append(p, *n.Name))
}

func (di *DocumentIndexer) addUsers(doc *Document, summaries *[]benchlingsdk.UserSummary) {
	if summaries == nil {
		return
	}
	for _, u := range *summaries {
		doc.Users[*u.Id] = di.users["user:"+*u.Id]
	}
}
