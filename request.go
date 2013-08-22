package crocodoc

import (
	"errors"
	"fmt"
	"github.com/fvbock/gorequests"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

type CrocoDoc struct {
	Uuid                string    `json:"uuid,omitempty"` // the identifier for the document on the crocodoc service
	Filename            string    `json:"-"`
	Status              int       `json:"-"`
	ExtractedText       string    `json:"-"`
	SessionId           string    `json:"session,omitempty"`
	SessionIdValidUntil time.Time `json:"-"`
}

func (c *CrocoDoc) String() string {
	return fmt.Sprintf("<CrocoDoc:: Id: %s, Status: %v, Filename: %s, SessionId(valid until %v): %s,  Text extracted? %v.>", c.Uuid, c.Status, c.Filename, c.SessionIdValidUntil, c.SessionId, len(c.ExtractedText) > 0)
}

/*
Upload takes an io stream from which it will copy data and upload it to crocodoc. The returned `CrocoDoc` will have th Uuid set - nothing else.
*/
func Upload(fs io.ReadCloser, filename string) (cf *CrocoDoc, err error) {
	data := map[string]string{
		"token": CrocoDocToken,
	}

	files := map[string]map[string]io.ReadCloser{
		"file": map[string]io.ReadCloser{filename: fs},
	}

	r := gorequests.Retry(gorequests.Post(UPLOAD_ENDPOINT, data, files, -1), MAX_RETRY_ON_RATELIMIT, RETRY_ON_RATELIMIT_TIMEOUT, []int{400})
	if r.Error != nil {
		log.Println(r.Error)
		return
	}
	err = checkResponse(r, false)
	if err != nil {
		log.Println(err)
		return
	}

	r.UnmarshalJson(&cf)
	cf.Filename = filename
	return
}

/*
UploadFile is a convinience wrapper for `Upload()` that takes a filename as arg to then pass the file handler to Upload()
*/
func UploadFile(filename string) (cf *CrocoDoc, err error) {
	fh, err := os.Open(filename)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error opening file(%s): %v", filename, err))
		log.Println(err)
		return
	}

	cf, err = Upload(fh, fh.Name())
	return
}

func (c *CrocoDoc) GetStatus() (err error) {
	if c.Uuid == "" {
		err = errors.New("Cannot call Status: No UUID is set on the CrocoDoc.")
		debug.PrintStack()
		return
	}
	statuslist, err := GetStatusesForIds([]string{c.Uuid})
	if err != nil {
		log.Println(err)
		return
	}

	switch statuslist[0].Status {
	case "QUEUED":
		c.Status = QUEUED
	case "PROCESSING":
		c.Status = PROCESSING
	case "DONE":
		c.Status = DONE
	case "ERROR":
		c.Status = ERROR
		err = errors.New(statuslist[0].Error)
	default:
		c.Status = ERROR
		err = errors.New(statuslist[0].Error)
	}
	return
}

/*
Delete removes the document from crocodoc
*/
func (c *CrocoDoc) Delete() (deleted bool, err error) {
	if c.Uuid == "" {
		err = errors.New("Cannot call Delete: No UUID is set on the CrocoDoc.")
		return
	}
	data := map[string]string{
		"token": CrocoDocToken,
		"uuid":  c.Uuid,
	}

	r := gorequests.Post(DELETE_ENDPOINT, data, nil, -1)
	if r.Error != nil {
		log.Println(err)
	}

	err = checkResponse(r, false)
	if err != nil {
		log.Println(err)
		return
	}

	err = r.UnmarshalJson(&deleted)
	if err != nil {
		log.Println(err)
	}

	if deleted {
		c.Status = DELETED
		c.SessionId = ""
		c.SessionIdValidUntil = time.Time{}
	}

	return
}

/*
CreateSession checks and if necessary reinitializes a viewing session for the document. The session id will be set on the local item.
*/
func (c *CrocoDoc) CreateSession() (err error) {
	if c.Uuid == "" {
		err = errors.New("Cannot call CreateSession: No UUID is set on the CrocoDoc.")
		return
	}
	if !c.SessionIdValidUntil.IsZero() && time.Now().Before(c.SessionIdValidUntil) {
		return
	}
	data := map[string]string{
		"token": CrocoDocToken,
		"uuid":  c.Uuid,
	}

	r := gorequests.Post(SESSION_ENDPOINT, data, nil, -1)
	if r.Error != nil {
		log.Println(r.Error)
		return
	}

	err = checkResponse(r, false)
	if err != nil {
		log.Println(err)
		return
	}

	err = r.UnmarshalJson(&c)
	if err != nil {
		log.Println(err)
		return
	}
	c.SessionIdValidUntil = time.Now().Add(SESSION_LIFETIME_MINUTES * time.Minute)
	return
}

/*
GetText retrieves the extracted text from the document from crocodoc and sets it on the local item.
*/
func (c *CrocoDoc) GetText() (err error) {
	if c.Uuid == "" {
		err = errors.New("Cannot call GetText: No UUID is set on the CrocoDoc.")
		return
	}
	if len(c.ExtractedText) > 0 {
		return
	}
	data, err := gorequests.NewQueryData(
		map[string]string{
			"token": CrocoDocToken,
			"uuid":  c.Uuid,
		})
	if err != nil {
		log.Println(err)
		return
	}

	r := gorequests.Get(GETTEXT_ENDPOINT, data, -1)
	if r.Error != nil {
		log.Println(r.Error)
		return
	}
	err = checkResponse(r, false)
	if err != nil {
		log.Println(err)
		return
	}

	// The text for each page is separated by the form feed character (U+000C)
	c.ExtractedText, err = r.Text()
	if err != nil {
		log.Println(err)
		return
	}
	c.ExtractedText = strings.TrimSpace(c.ExtractedText)
	return
}

/*
Download downloads either an office file or a PDF version of it from crocodoc and saves the file locally. If an empty filename is proviced it will use the <UUID.png> as the name pattern.

The parameters `withAnnotations` adn `filterUserAnnotations` work the same way as described in the crocodoc API reference:

	annotated - Include annotations. If true, downloaded document will be a PDF.

	Default: false

	filter - Limit which users' annotations included. Possible values are:
	all, none, or a comma-separated list of user IDs as supplied in the user
	field when creating sessions. See the filter parameter of session
	creation for example values.

	Default: all
*/
func (c *CrocoDoc) Download(asPdf bool, filename string, withAnnotations bool, filterUserAnnotations string) (err error) {
	if c.Uuid == "" {
		err = errors.New("Cannot call Download: No UUID is set on the CrocoDoc.")
		return
	}

	var renameTo string
	/* crocodoc only allows certain characters as filename in the request.
	if the filename provided is not valid we will use the uuid as an intermediate filename for the request and then save it with the provided name
	*/
	if !allowedFilename(filename) {
		renameTo = filename
		filename = c.Uuid
	}

	data, err := gorequests.NewQueryData(
		map[string]string{
			"token":     CrocoDocToken,
			"uuid":      c.Uuid,
			"pdf":       asString(asPdf),
			"filename":  filename,
			"annotated": asString(withAnnotations),
			"filter":    filterUserAnnotations,
		})
	if err != nil {
		log.Println(err)
	}

	r := gorequests.Get(DOWNLOAD_ENDPOINT, data, -1)
	log.Println(r.Headers())
	if r.Error != nil {
		log.Println(r.Error)
	}

	err = checkResponse(r, false)
	if err != nil {
		log.Println(err)
		return
	}

	if len(renameTo) == 0 {
		err = r.IntoFile(fileLocation(filename))
	} else {
		err = r.IntoFile(fileLocation(renameTo))
	}

	if err != nil {
		log.Println(err)
	}

	return
}

/*
Thumbnail gets a thumbnail from crocodoc in the given size (max '300x300') and saves the file locally. If an empty filename is proviced it will use the <UUID.png> as the name pattern.
*/
func (c *CrocoDoc) Thumbnail(size string, filename string) (err error) {
	if c.Uuid == "" {
		err = errors.New("Cannot call Thumbnail: No UUID is set on the CrocoDoc.")
		return
	}

	if len(size) == 0 {
		size = DEFAULT_THUMBNAIL_SIZE
	}
	if strings.Index("x", size) != -1 {
		err = errors.New("Error: Thumbnail size needs to be specified as a string of the from '100x100'. (maximum size is '300x300')")
		return
	} else {
		dimensions := strings.Split(size, "x")
		x, errx := strconv.ParseInt((dimensions[0]), 10, 0)
		y, erry := strconv.ParseInt((dimensions[1]), 10, 0)
		if errx != nil || erry != nil || x < 1 || x > 300 || y < 1 || y > 300 {
			err = errors.New("Error: Thumbnail size needs to be specified as a string of the from '100x100'. (maximum size is '300x300')")
		}
	}

	data, err := gorequests.NewQueryData(
		map[string]string{
			"token": CrocoDocToken,
			"uuid":  c.Uuid,
			"size":  size,
		})
	if err != nil {
		log.Println(err)
	}

	r := gorequests.Get(THUMBNAIL_ENDPOINT, data, -1)
	log.Println(r.Headers())
	if r.Error != nil {
		log.Println(r.Error)
	}

	err = checkResponse(r, false)
	if err != nil {
		log.Println(err)
		return
	}

	if len(filename) == 0 {
		err = r.IntoFile(fileLocation(fmt.Sprintf("%s.png", c.Uuid)))
	} else {
		err = r.IntoFile(fileLocation(filename))
	}

	if err != nil {
		log.Println(err)
	}

	return
}

type StatusResponse struct {
	Uuid     string `json:"uuid"`
	Status   string `json:"status"`
	Viewable bool   `json:"viewable"`
	Error    string `json:"error,omitempty"`
}

func (s *StatusResponse) String() string {
	return fmt.Sprintf("Id: %s, Status: %s, Viewable: %v, Error: %s.", s.Uuid, s.Status, s.Viewable, s.Error)
}

/*
GetStatusesForIds retrieves `StatusResponse`s for each crocodoc UUID given.
*/
func GetStatusesForIds(uuids []string) (statuslist []*StatusResponse, err error) {
	if len(uuids) == 0 {
		err = errors.New("Cannot call GetStatusesForIds: Need at least one UUID to be set.")
		return
	}
	data, err := gorequests.NewQueryData(
		map[string]string{
			"token": CrocoDocToken,
			"uuids": strings.Join(uuids, ","),
		})
	if err != nil {
		log.Println(err)
		return statuslist, err
	}

	r := gorequests.Get(STATUS_ENDPOINT, data, -1)
	if r.Error != nil {
		log.Println(r.Error)
		return statuslist, r.Error
	}
	err = checkResponse(r, false)
	if err != nil {
		log.Println(err)
		return
	}

	err = r.UnmarshalJson(&statuslist)
	if err != nil {
		log.Println(err)
	}
	return
}
