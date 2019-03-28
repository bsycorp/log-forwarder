package main

import (
	"github.com/coreos/go-systemd/sdjournal"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

// Block for up to this long waiting for a journal entry
const JournalWaitTime  = time.Duration(2) * time.Second

type JournalReader struct {
	StateFilePath string
	Cursor string
	Journal *sdjournal.Journal
}

func (jr *JournalReader) Open(stateFilePath string) {
	jr.StateFilePath = stateFilePath
	jr.Cursor = jr.readStateFile()
	jr.openJournal()
}

func (jr *JournalReader) GetNextEntry () (*sdjournal.JournalEntry) {
	// Look for a new Journal entry
	r, err := jr.Journal.Next()
	if err != nil {
		log.Fatalln("Error getting next journal entry: ", err)
	}
	if r == 0 { // Nothing right now
		// We're at the end of the journal, so wait on it for a short while
		// Maybe something will turn up!
		waitResult := jr.Journal.Wait(JournalWaitTime)

		// So one of 3 things can happen next:
		// SD_JOURNAL_NOP        - The journal did not change (timeout)
		// SD_JOURNAL_APPEND     - New journal entries are available
		// SD_JOURNAL_INVALIDATE - Journal files were added or removed

		if waitResult != sdjournal.SD_JOURNAL_APPEND {
			return nil
		}
	}

	// OK, if we made it here there is supposed to be a journal
	// entry waiting for us
	ent, err := jr.Journal.GetEntry()
	if err != nil {
		// We sometimes see errors like: "failed to get realtime timestamp: 99"
		// from sdjournal, a.k.a EADDRNOTAVAIL.
		// See also e.g: https://github.com/systemd/systemd-netlogd/pull/15/files
		// We try re-opening the journal to fix this.
		if strings.Contains(err.Error(), "timestamp: 99") {
			log.Println("warning: GetEntry: ", err)
			jr.openJournal()
			return nil
		} else {
			log.Fatalln("fatal: GetEntry: ", err)
		}
	}

	jr.Cursor = ent.Cursor
	jr.writeStateFile()
	return ent
}

func (jr *JournalReader) openJournal() {
	log.Println("Opening journal")

	j, err := sdjournal.NewJournal()
	if err != nil {
		log.Fatalln("Could not open journal:", err)
	}
	jr.Journal = j

	// Prevent auto-truncation of log entries at 64k
	err = jr.Journal.SetDataThreshold(0)
	if err != nil {
		log.Fatalln("Could not set journal data threshold:", err)
	}

	if jr.Cursor == "" {
		log.Println("No last cursor, starting from beginning")
	} else {
		log.Println("Seeking to: ", jr.Cursor)
		err = jr.Journal.SeekCursor(jr.Cursor)
		if err != nil {
			log.Fatalln("Error seeking to cursor: ", err)
		}
	}
}

func (jr *JournalReader) readStateFile() string {
	b, err := ioutil.ReadFile(jr.StateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("State file not found")
			return ""
		}
		log.Fatalln("Failed to open state file: ", err)
	}
	return string(b)
}

func (jr *JournalReader) writeStateFile() {
	err := ioutil.WriteFile(jr.StateFilePath, []byte(jr.Cursor), os.FileMode(0600))
	if err != nil {
		log.Fatalln("Failed to write state file:", err)
	}
}
