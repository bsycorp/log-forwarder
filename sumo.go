package main

import (
	//"github.com/pkg/errors"
	"log"
	"net/http"
	"os"

	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

type SumoUploader struct {
	httpClient *http.Client

	//this is the url of a sumo http collector with 'Enabled timestamp parsing' ON
	//so it will try and parse timestamp in log messages. We can only send messages
	//to this endpoint if we 'trust' the format so they get processed correctly
	TrustedTimestampCollectorUrl string

	//this is the url of a sumo http collector with 'Enabled timestamp parsing' OFF
	//so it will just use the receipt time / processing time of the log entry as the
	//searchable timestamp. This is the least worst way of making log entry timing mostly correct
	UntrustedTimestampCollectorUrl string
}

//Must get a value from one of the given env variables or fail
func MustGetEnv(envVariables ...string) string {
	for _, envValue := range envVariables {
		foundValue := os.Getenv(envValue)
		if foundValue != "" {
			return foundValue
		}
	}

	//fail
	log.Fatalf("Required environment variable not set: %s", envVariables)
	return ""
}

/*
	See: https://help.sumologic.com/Send-Data/Sources/02Sources-for-Hosted-Collectors/HTTP-Source

	SumoLogic Headers:

	X-Sumo-Name: Desired source name.
	X-Sumo-Host: Desired host name.
	X-Sumo-Category: Desired source category.

	SumoLogic Response Codes:

	200	HTTP request received and processed successfully.
	401	HTTP request was rejected due to missing or invalid URL token.
	408	HTTP request was accepted, but timed out processing. For more information, see Request Timeouts.
	429	HTTP request was rejected due to quota-based throttling. For more information, see Throttling, below.
	503	HTTP request was rejected due to server issues. Check the Status page in the Sumo web app
	504	HTTP request was rejected due to server issue. Check the Status page in the Sumo web app
*/

func backoff(n int) int {
	delay := []int{0, 1, 2, 4, 8, 16, 32, 64}
	if n >= len(delay) {
		return 120
	} else {
		return delay[n]
	}
}

func compress(s string) []byte {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(s)); err != nil {
		panic(err)
	}
	if err := gz.Flush(); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func (sumo *SumoUploader) UploadLogEntries(metadata MetadataValues, lines []string) {
	const lineSep = "\n"
	const requestTimeout = 10 * time.Second

	logData := compress(strings.Join(lines, lineSep))

Retry:
	for attempts := 0; ; /* no condition... */ attempts++ {
		backoffSecs := backoff(attempts)
		if backoffSecs > 0 {
			log.Printf("Backing off for %d seconds", backoffSecs)
			time.Sleep(time.Duration(backoffSecs) * time.Second)
		}
		// ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
		// TODO: should not be willing to block forever here.

		uploadStart := time.Now()

		collectorURL := sumo.TrustedTimestampCollectorUrl
		if metadata.trustedTimestamp == false {
			collectorURL = sumo.UntrustedTimestampCollectorUrl
		}

		req, err := http.NewRequest("POST", collectorURL, bytes.NewReader(logData))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("X-Sumo-Name", metadata.source)
		req.Header.Set("X-Sumo-Host", metadata.host)
		req.Header.Set("X-Sumo-Category", metadata.category)
		req.Header.Set("Content-Encoding", "gzip")

		resp, err := sumo.httpClient.Do(req)
		if err != nil {
			log.Println("Error uploading logs:", err)
			continue Retry
		}
		defer resp.Body.Close()
		_, err = io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			log.Println("Error reading sumo response:", err)
			continue Retry
		}
		if resp.StatusCode != 200 {
			log.Println("Failed upload to sumo server, status code: ", resp.StatusCode)
			continue Retry
		}

		// We did it ┣┓웃┏♨❤♨┑유┏┥
		uploadTook := time.Since(uploadStart)
		log.Printf("Uploaded %d bytes in %s", len(logData), uploadTook.String())
		break
	}
}
