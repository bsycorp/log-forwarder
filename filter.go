package main

import "github.com/coreos/go-systemd/sdjournal"

// Filtering for log messages
// See: https://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html
// For expected fields in Journal Entries

type FilterFn func(e *sdjournal.JournalEntry) bool

type FilterChain struct {
	funcs []FilterFn
}

// Add a filter function to the filter chain
func (fs *FilterChain) AddFilter(fn FilterFn) {
	fs.funcs = append(fs.funcs, fn)
}

// Returns true iff all the filters in the stack return true
func (stack *FilterChain) Want(e *sdjournal.JournalEntry) bool {
	for _, fn := range stack.funcs {
		if !fn(e) {
			return false
		}
	}
	return true
}

// Returns a filter that accepts log entries iff the transport of
// the journal entry is in desiredTransports
func FilterByTransport(desiredTransports []string) FilterFn {
	return func(e *sdjournal.JournalEntry) bool {
		return ListContains(desiredTransports, e.Fields["_TRANSPORT"])
	}
}

//Returns a filter that filters out specific service / unit names, useful to exclude the logs of the logfwder itself
func ExcludeBySystemDUnit(excludedUnits []string) FilterFn {
	return func(e *sdjournal.JournalEntry) bool {
		//if we have a unit, and its an excluded one then filter, allow everything else.
		return !ListContains(excludedUnits, e.Fields["_SYSTEMD_UNIT"])
	}
}
