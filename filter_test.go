package main

import (
	"github.com/coreos/go-systemd/sdjournal"
	"github.com/stretchr/testify/assert"
	"testing"
)

var msg1 = sdjournal.JournalEntry{
	Fields: map[string]string{
		"MESSAGE":                     "Linux version 4.14.32-coreos (jenkins@jenkins-worker-7) (gcc version 7.3.0 (Gentoo Hardened 7.3.0 p1.0)) #1 SMP Tue Apr 3 05:21:26 UTC 2018",
		"PRIORITY":                    "5",
		"SYSLOG_FACILITY":             "0",
		"SYSLOG_IDENTIFIER":           "kernel",
		"_BOOT_ID":                    "c3c4787f604643e7b030a8d6420bbe7e",
		"_HOSTNAME":                   "localhost",
		"_MACHINE_ID":                 "26fae286b7f84ea5b91cde60472b2590",
		"_SOURCE_MONOTONIC_TIMESTAMP": "0",
		"_TRANSPORT":                  "kernel",
	},
	Cursor:             "s=5a801c1c37e043eea7b4cfd89aca4584;i=1;b=c3c4787f604643e7b030a8d6420bbe7e;m=182334;t=56ce74eabceeb;x=d74f686daf0e970d",
	RealtimeTimestamp:  1527115596680939,
	MonotonicTimestamp: 1581876,
}
var msg2 = sdjournal.JournalEntry{
	Fields: map[string]string{
		"CODE_FILE":                  "../systemd-237/src/sysctl/sysctl.c",
		"CODE_FUNC":                  "apply_all",
		"CODE_LINE":                  "61",
		"ERRNO":                      "2",
		"MESSAGE":                    "Couldn't write 'fq_codel' to 'net/core/default_qdisc', ignoring: No such file or directory",
		"PRIORITY":                   "5",
		"SYSLOG_FACILITY":            "3",
		"SYSLOG_IDENTIFIER":          "systemd-sysctl",
		"_BOOT_ID":                   "c3c4787f604643e7b030a8d6420bbe7e",
		"_GID":                       "0",
		"_HOSTNAME":                  "localhost",
		"_MACHINE_ID":                "26fae286b7f84ea5b91cde60472b2590",
		"_PID":                       "145",
		"_SELINUX_CONTEXT":           "kernel",
		"_SOURCE_REALTIME_TIMESTAMP": "1527115596731568",
		"_TRANSPORT":                 "journal",
		"_UID":                       "0",
	},
	Cursor:             "s=5a801c1c37e043eea7b4cfd89aca4584;i=177;b=c3c4787f604643e7b030a8d6420bbe7e;m=18e8e9;t=56ce74eac949f;x=ce148813d1c90e32",
	RealtimeTimestamp:  1527115596731551,
	MonotonicTimestamp: 1632489,
}
var msg3 = sdjournal.JournalEntry{
	Fields: map[string]string{
		"MESSAGE":                     "Linux version 4.14.32-coreos (jenkins@jenkins-worker-7) (gcc version 7.3.0 (Gentoo Hardened 7.3.0 p1.0)) #1 SMP Tue Apr 3 05:21:26 UTC 2018",
		"PRIORITY":                    "5",
		"_SYSTEMD_UNIT":               "test.service",
		"_BOOT_ID":                    "c3c4787f604643e7b030a8d6420bbe7e",
		"_HOSTNAME":                   "localhost",
		"_MACHINE_ID":                 "26fae286b7f84ea5b91cde60472b2590",
		"_SOURCE_MONOTONIC_TIMESTAMP": "0",
		"_TRANSPORT":                  "kernel",
	},
	Cursor:             "s=5a801c1c37e043eea7b4cfd89aca4584;i=1;b=c3c4787f604643e7b030a8d6420bbe7e;m=182334;t=56ce74eabceeb;x=d74f686daf0e970d",
	RealtimeTimestamp:  1527115596680939,
	MonotonicTimestamp: 1581876,
}

func TestFilterChain(t *testing.T) {
	fc := FilterChain{}

	// If there are no filters then we keep by default
	assert.True(t, fc.Want(&msg1))
	assert.True(t, fc.Want(&msg2))

	fc.AddFilter(FilterByTransport([]string{"kernel", "syslog", "stdout"}))
	assert.True(t, fc.Want(&msg1))
	assert.False(t, fc.Want(&msg2))

	fc = FilterChain{}

	assert.True(t, fc.Want(&msg1))
	assert.True(t, fc.Want(&msg2))

	fc.AddFilter(ExcludeBySystemDUnit([]string{"test.service"}))
	assert.True(t, fc.Want(&msg1))
	assert.True(t, fc.Want(&msg2))
	assert.False(t, fc.Want(&msg3))
}
