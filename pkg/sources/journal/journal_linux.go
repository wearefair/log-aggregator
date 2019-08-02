// Package journal for linux
// These methods require being compiled on a system with the systemd headers
package journal

import (
	"time"

	"github.com/cenkalti/backoff"
	"github.com/coreos/go-systemd/sdjournal"
	"github.com/pkg/errors"
	"github.com/wearefair/log-aggregator/pkg/logging"
	"github.com/wearefair/log-aggregator/pkg/types"
)

type JournalEntry sdjournal.JournalEntry

// Fields to strip from the entry prior to sending out as a record.
var omitFields = []string{
	sdjournal.SD_JOURNAL_FIELD_CURSOR,
	sdjournal.SD_JOURNAL_FIELD_MONOTONIC_TIMESTAMP,
	sdjournal.SD_JOURNAL_FIELD_BOOT_ID,
	sdjournal.SD_JOURNAL_FIELD_UID,
	sdjournal.SD_JOURNAL_FIELD_GID,
	sdjournal.SD_JOURNAL_FIELD_CAP_EFFECTIVE,
	sdjournal.SD_JOURNAL_FIELD_SYSTEMD_SLICE,
	sdjournal.SD_JOURNAL_FIELD_SYSLOG_IDENTIFIER,
	sdjournal.SD_JOURNAL_FIELD_SYSTEMD_CGROUP,
	sdjournal.SD_JOURNAL_FIELD_CMDLINE,
	sdjournal.SD_JOURNAL_FIELD_COMM,
	sdjournal.SD_JOURNAL_FIELD_SELINUX_CONTEXT,
	sdjournal.SD_JOURNAL_FIELD_SYSLOG_FACILITY,
	sdjournal.SD_JOURNAL_FIELD_REALTIME_TIMESTAMP,
	sdjournal.SD_JOURNAL_FIELD_SOURCE_REALTIME_TIMESTAMP,
	sdjournal.SD_JOURNAL_FIELD_PRIORITY,
	sdjournal.SD_JOURNAL_FIELD_TRANSPORT,
	sdjournal.SD_JOURNAL_FIELD_MACHINE_ID,
	sdjournal.SD_JOURNAL_FIELD_EXE,
	sdjournal.SD_JOURNAL_FIELD_HOSTNAME,
}

type Client struct {
	shutdown bool
	out      chan<- *types.Record
	journal  *sdjournal.Journal
}

func New(conf ClientConfig) (client *Client, err error) {
	var journal *sdjournal.Journal
	if conf.JournalDirectory == "" {
		journal, err = sdjournal.NewJournal()
	} else {
		journal, err = sdjournal.NewJournalFromDir(conf.JournalDirectory)
	}
	if err != nil {
		return nil, errors.Wrap(err, "Error constructing systemd Journal client")
	}

	if string(conf.Cursor) != "" {
		err = journal.SeekCursor(string(conf.Cursor))
		if err != nil {
			return nil, errors.Wrapf(err, "Error seeking to cursor %s", conf.Cursor)
		}
		// The cursor positions us on the previously read item, so advance to the next one (if possible).
		_, err = journal.Next()
		if err != nil {
			return nil, errors.Wrap(err, "Error advancing to next entry after seeking to cursor")
		}
	}
	return &Client{
		journal: journal,
	}, nil
}

func (c *Client) Start(out chan<- *types.Record) {
	c.out = out
	go c.read()
}

func (c *Client) read() {
	var entry *sdjournal.JournalEntry
	var count uint64
	var err error

	for !c.shutdown {
		// If the error is not nil from the previous run, sleep for half a second
		if err != nil {
			time.Sleep(time.Millisecond * 500)
		}
		count, err = c.journal.Next()
		if err != nil {
			logging.Error(errors.Wrap(err, "Got error advancing entry from systemd Journal"))
			continue
		}
		if count == 0 {
			// Wait for new journal events
			c.journal.Wait(time.Second * 5)
			continue
		}
		// If reading the entry fails (we have already retried)
		// then panic, as there is no way to recover
		entry, err = c.readEntry()
		if err != nil {
			logging.Error(err)
			panic(err)
		}

		c.out <- entryToRecord((*JournalEntry)(entry))
	}
}

func (c *Client) readEntry() (entry *sdjournal.JournalEntry, err error) {
	readHelper := func() error {
		entry, err = c.journal.GetEntry()
		return err
	}
	// Call once before setting up retry logic
	readHelper()
	if err == nil {
		return entry, nil
	}

	strategy := backoff.NewExponentialBackOff()
	strategy.MaxElapsedTime = time.Second * 15
	err = backoff.Retry(readHelper, strategy)
	if err != nil {
		return nil, errors.Wrap(err, "Got error reading entry from systemd Journal")
	}
	return entry, nil
}
