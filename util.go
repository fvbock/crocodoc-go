package crocodoc

import (
	"errors"
	"fmt"
	"gorequests"
	"strings"
)

type CrocoError struct {
	Message string `json:"error,omitempty"`
}

/*
CheckResponse checks errors/messages for NON-ok (!= 200) requests.

Errors that might be inside a status 200 response (eg. statuses requests) need
to be checked separately.
*/
func CheckResponse(r *gorequests.Response, noJson bool) (err error) {
	if r.Status != 200 {
		var jsonErrorMsg string
		if !noJson {
			var ce *CrocoError
			err = r.UnmarshalJson(&ce)
			if len(ce.Message) > 0 {
				jsonErrorMsg = fmt.Sprintf("CrocoDoc Error (HTTP status %v): %s", r.Status, ce.Message)
			}

		}

		if r.Status >= 500 && r.Status < 600 {
			err = errors.New(fmt.Sprintf("Unknown server error: status %v", r.Status))
			return
		} else {
			for s, msg := range Http_4xx_errors {
				if s == r.Status {
					err = errors.New(fmt.Sprintf("Server_error_%v: %s. %s", r.Status, msg, jsonErrorMsg))
					break
				}
			}
		}
	}
	return
}

func asString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func allowedFilename(s string) bool {
	// filenames must use chars from space (0x20) up to ~ (0x7E)
	for _, c := range s {
		if c < 0x20 || c > 0x7E {
			return false
		}
	}
	return true
}

func fileLocation(filename string) string {
	if strings.Index("/", filename) != -1 {
		return fmt.Sprintf("%s/%s", DEFAULT_FILE_PATH, filename)
	} else {
		return filename
	}
}
