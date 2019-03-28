package main

import (
	"flag"
	"fmt"
	"github.com/coreos/go-systemd/sdjournal"
	"github.com/patrickmn/go-cache"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	DefaultStateFile = "log-forwarder.state"
)

var stateFile = flag.String("statefile", DefaultStateFile, "File to checkpoint log position for resuming.")
var metricsArg = flag.String("metrics", "none", "metrics provider (none,datadog,prometheus)")

const activeBufferExpiry = 24*time.Hour
const seenCursorExpiry = 10*time.Minute


//Maintain an expiring map of queues, will look up by queue identifier, inactive queues will be collected by the map
var activeBuffers = cache.New(activeBufferExpiry, activeBufferExpiry)

//This is for debugging; lets us know if for we're reprocessing log messages for whatever reason.
var seenCursors = cache.New(seenCursorExpiry, seenCursorExpiry)

func main() {
	flag.Parse()

	metrics := &Metrics{}
	metrics.Init()

	sumoUploader := &SumoUploader{
		httpClient:                     &http.Client{},
		Metrics:                        metrics,
		TrustedTimestampCollectorUrl:   MustGetEnv("SUMO_TRUSTED_TIMESTAMP_COLLECTOR_URL", "SUMO_COLLECTOR_URL"),
		UntrustedTimestampCollectorUrl: MustGetEnv("SUMO_UNTRUSTED_TIMESTAMP_COLLECTOR_URL", "SUMO_COLLECTOR_URL"),
	}

	//setup metadata defaults
	SetMetadataDefaults(MetadataValues{
		MustGetEnv("SUMO_SOURCE_NAME"),
		MustGetEnv("SUMO_SOURCE_CATEGORY"),
		GetHostname(os.Getenv("SUMO_SOURCE_HOST")),
		true,
	})

	jr := &JournalReader{}
	jr.Open(*stateFile)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	eventFilters := &FilterChain{}
	allowedEventTransports := MakeTransportList(
		Split(os.Getenv("JOURNAL_INCLUDE_TRANSPORTS"), ","),
		Split(os.Getenv("JOURNAL_EXCLUDE_TRANSPORTS"), ","),
	)
	log.Println("Listening for journald transports: ", allowedEventTransports)
	eventFilters.AddFilter(FilterByTransport(allowedEventTransports))

	formatMessageFilters := &FilterChain{}
	excludeFormatMessageUnits := Split(os.Getenv("FORMAT_MESSAGE_EXCLUDE_UNITS"), ",")
	if len(excludeFormatMessageUnits) == 0 {
		//add docker by default if not specified
		excludeFormatMessageUnits = []string{"docker.service"}
	}
	log.Println("Not formatting message for systemd units: ", excludeFormatMessageUnits)
	formatMessageFilters.AddFilter(ExcludeBySystemDUnit(excludeFormatMessageUnits))

	excludeUnits := Split(os.Getenv("JOURNAL_EXCLUDE_UNITS"), ",")
	if len(excludeUnits) > 0 {
		eventFilters.AddFilter(ExcludeBySystemDUnit(excludeUnits))
		log.Println("Excluding systemd units named: ", excludeUnits)
	}

	excludeSumoCategories := Split(os.Getenv("SUMO_EXCLUDE_SOURCE_CATEGORIES"), ",")
	if len(excludeSumoCategories) > 0 {
		log.Println("Excluding messages for sumo source categories: ", excludeSumoCategories)
	}

	metrics.Start(*metricsArg)
	var mainLoopLast time.Time
	var lastCursor string

MainLoop:
	for {
		metrics.MainLoopSpins.Inc(1)
		if !mainLoopLast.IsZero() {
			metrics.MainLoopTime.UpdateSince(mainLoopLast)
		}
		mainLoopLast = time.Now()

		// Non-blocking check for SIGINT or SIGTERM
		select {
		case _ = <-sigCh:
			break MainLoop
		default:
		}

		ent := jr.GetNextEntry()
		if ent != nil {
			if ent.Cursor == lastCursor {
				// Hrm, same cursor? Ok, skip it
				metrics.DebugSkippedCursor.Inc(1)
			} else if eventFilters.Want(ent) {
				//by default just use the raw message
				logMessage := ent.Fields["MESSAGE"]

				//optionally, if the transport is configured to be formatted then use a formatted message instead
				if formatMessageFilters.Want(ent) {
					logMessage = FormatLogEntry(ent)
				}

				//lookup correct buffer for entry
				buf := getOrCreateActiveBufferForEntry(ent)

				//check whether category is excluded
				if !isSumoCategoryExcluded(buf.Metadata.category, excludeSumoCategories) {
					//append desired msg to that queue
					buf.Append(logMessage)
					err := seenCursors.Add(ent.Cursor, nil, seenCursorExpiry)
					if err != nil {
						// Shouldn't happen!
						log.Println("Error: processing previously seen cursor: ", ent.Cursor)
						metrics.DebugDupCursor.Inc(1)
					}
				}
			}
			lastCursor = ent.Cursor
		}

		//loop through buffers, start goroutine to check flush, do upload if required and then clear buffer.
		activeBufferItems := activeBuffers.Items()
		metrics.BuffersActive.Update(int64(len(activeBufferItems)))

		uploadCh := make(chan int)
		for _, item := range activeBufferItems {
			go func(bufferItem cache.Item, uploadCh chan<- int) {
				buf := bufferItem.Object.(*LogBuffer)
				if buf.NeedsFlush() {
					sumoUploader.UploadLogEntries(buf.Metadata, buf.GetMessages())
					//done uploading, so clear sent buffer
					buf.Clear()
				}
				uploadCh <- 0
			}(item, uploadCh)
		}

		// Wait for each of the buffer goroutines flush to finish, fire goroutine for every item even if its doesn't
		// need to flush, easier to know when we are finished that way,
		for range activeBufferItems {
			// We can block here for some time (actually will wait indefinitely
			// for the buffer to upload to Sumo) so we do that in the background
			// and check for shutdown signals while we wait.
			select {
			case _ = <-sigCh: // We got SIGINT or SIGTERM
				break MainLoop
			case _ = <-uploadCh: // We successfully uploaded to Sumo
			}
		}

		close(uploadCh)
	}

	// Don't bother cleaning up or flushing anything.
	log.Println("Caught signal, shutting down.")
}

func FormatLogEntry(ent *sdjournal.JournalEntry) string {
	ts := time.Unix(0, int64(ent.RealtimeTimestamp*1000)).Format(time.RFC3339)
	return fmt.Sprintf("%s: %s: %s", ts, ent.Fields["_HOSTNAME"], ent.Fields["MESSAGE"])
}

func getOrCreateActiveBufferForEntry(ent *sdjournal.JournalEntry) *LogBuffer {
	bufferIdentifier := getLogBufferIdentifierForEntry(ent)
	buffer, found := activeBuffers.Get(getLogBufferIdentifierForEntry(ent))
	if !found {
		//if not found then create a new log buffer, setup metadata and cache it
		metadata := getMetadataForLogEntry(ent)
		buffer = &LogBuffer{
			Metadata: metadata,
		}
		err := activeBuffers.Add(bufferIdentifier, buffer, activeBufferExpiry)
		if err != nil {
			log.Fatalln("Error creating log buffer for: ", bufferIdentifier, err)
		}
	}

	return buffer.(*LogBuffer)
}

func isSumoCategoryExcluded(category string, excludedCategories []string) bool {
	for _, ex := range excludedCategories {
		if strings.Contains(category, ex) {
			return true
		}
	}
	return false
}

//returns value representing the correct queue for this entry, used to separate different entry types so they can have different metadata
func getLogBufferIdentifierForEntry(ent *sdjournal.JournalEntry) string {
	//if the entry is from systemd, its either a container from docker or another systemd service
	if len(ent.Fields["_SYSTEMD_SLICE"]) > 0 {

		if len(ent.Fields["CONTAINER_ID"]) > 0 {
			//if we have a containerID then the entry is from docker (or something like it? assume docker for now)
			return "docker-" + ent.Fields["CONTAINER_ID"]
		} else {
			//if no container id, then its systemd but its not docker, so just use the systemd unit name as the value
			return "systemd-" + ent.Fields["SYSLOG_IDENTIFIER"]
		}

	} else {
		//where its not systemd or docker, then just separate by transport (audit, stdout, kernel etc)
		return "journald-" + ent.Fields["_TRANSPORT"]
	}
}

func getMetadataForLogEntry(ent *sdjournal.JournalEntry) MetadataValues {
	//start with the defaults
	metadataValues := GetMetadataDefaults()

	//if the entry is from systemd, its either a container from docker or another systemd service
	if len(ent.Fields["_SYSTEMD_SLICE"]) > 0 {

		//if we have a containerID then the entry is from docker (or something like it? assume docker for now)
		if len(ent.Fields["CONTAINER_ID_FULL"]) > 0 {
			metadataValues = GetMetadataForContainerID(ent.Fields["CONTAINER_ID_FULL"])
		} else {
			metadataValues = GetMetadataForProcess("systemd", ent.Fields["SYSLOG_IDENTIFIER"])
		}
	} else {
		metadataValues = GetMetadataForProcess("journald", ent.Fields["_TRANSPORT"])
	}

	return metadataValues
}



func MakeTransportList(include []string, exclude []string) []string {
	validTransports := []string{"audit", "driver", "syslog", "journal", "stdout", "kernel"}
	var transports []string
	if len(include) > 0 {
		transports = ListIntersect(validTransports, include)
	} else {
		transports = validTransports
	}
	if len(exclude) > 0 {
		transports = ListSubtract(transports, exclude)
	}
	return transports
}
