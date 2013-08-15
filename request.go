package crocodoc

import (
	"errors"
	"fmt"
	"gorequests"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

type CrocoDoc struct {
	Uuid                string    `json:"uuid,omitempty"`
	Filename            string    `json:"-"`
	Status              int       `json:"-"`
	ExtractedText       string    `json:"-"`
	SessionId           string    `json:"session,omitempty"`
	SessionIdValidUntil time.Time `json:"-"`
}

func (c *CrocoDoc) String() string {
	return fmt.Sprintf("<CrocoDoc:: Id: %s, Status: %v, Filename: %s, SessionId(valid until %v): %s,  Text extracted? %v.>", c.Uuid, c.Status, c.Filename, c.SessionIdValidUntil, c.SessionId, len(c.ExtractedText) > 0)
}

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

// func Upload(filename string) (cf *CrocoDoc, err error) {
func Upload(fs io.ReadCloser, filename string) (cf *CrocoDoc, err error) {
	data := map[string]string{
		"token": CROCODOC_API_TOKEN,
	}

	files := map[string]map[string]io.ReadCloser{
		"file": map[string]io.ReadCloser{filename: fs},
	}

	r, err := gorequests.Post(UPLOAD_ENDPOINT, data, files, -1)
	if err != nil {
		log.Println(err)
		return
	}
	err = CheckResponse(&r, false)
	if err != nil {
		// TODO? retry w limit?
		// request_test.go:34: CrocoDoc Error (HTTP status 400): rate limit exceeded
		log.Println(err)
		return
	}

	r.UnmarshalJson(&cf)
	cf.Filename = filename
	log.Println(cf)
	return
}

// {"status": "DONE", "viewable": true, "uuid": "a1384501-3c5e-44e6-b26e-e9a0bbbef6e4"}
type StatusResponse struct {
	Uuid     string `json:"uuid"`
	Status   string `json:"status"`
	Viewable bool   `json:"viewable"`
	Error    string `json:"error,omitempty"`
}

func (s *StatusResponse) String() string {
	return fmt.Sprintf("Id: %s, Status: %s, Viewable: %v, Error: %s.", s.Uuid, s.Status, s.Viewable, s.Error)
}

func GetStatusesForIds(uuids []string) (statuslist []*StatusResponse, err error) {
	if len(uuids) == 0 {
		err = errors.New("Cannot call GetStatusesForIds: Need at least one UUID to be set.")
		return
	}
	data, err := gorequests.NewQueryData(
		map[string]string{
			"token": CROCODOC_API_TOKEN,
			"uuids": strings.Join(uuids, ","),
		})
	if err != nil {
		log.Println(err)
		return statuslist, err
	}

	r, err := gorequests.Get(STATUS_ENDPOINT, data, -1)
	if err != nil {
		log.Println(err)
		return statuslist, err
	}
	err = CheckResponse(&r, false)
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

func (c *CrocoDoc) GetStatus() (s *StatusResponse, err error) {
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

	s = statuslist[0]
	switch s.Status {
	case "QUEUED":
		c.Status = QUEUED
	case "PROCESSING":
		c.Status = PROCESSING
	case "DONE":
		c.Status = DONE
	case "ERROR":
		c.Status = ERROR
		err = errors.New(s.Error)
	default:
		c.Status = ERROR
		err = errors.New(s.Error)
	}
	return
}

// TODO:
// 2013/06/12 18:23:16 {"error": "invalid document uuid"}
// 2013/06/12 18:23:16 0
// 2013/06/12 18:23:16 application/json
// 2013/06/12 18:23:16 json decoding error: json: cannot unmarshal object into Go value of type bool

func (c *CrocoDoc) Delete() (deleted bool, err error) {
	if c.Uuid == "" {
		err = errors.New("Cannot call Delete: No UUID is set on the CrocoDoc.")
		return
	}
	data := map[string]string{
		"token": CROCODOC_API_TOKEN,
		"uuid":  c.Uuid,
	}

	r, err := gorequests.Post(DELETE_ENDPOINT, data, nil, -1)
	if err != nil {
		log.Println(err)
	}

	err = CheckResponse(&r, false)
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

// func (c *CrocoDoc) CreateSession(editable bool, user []string, filter []string, admin bool, downloadable bool, copyprotected bool, sidebar string) (err error) {
func (c *CrocoDoc) CreateSession() (err error) {
	if c.Uuid == "" {
		err = errors.New("Cannot call CreateSession: No UUID is set on the CrocoDoc.")
		return
	}
	if !c.SessionIdValidUntil.IsZero() && time.Now().Before(c.SessionIdValidUntil) {
		return
	}
	data := map[string]string{
		"token": CROCODOC_API_TOKEN,
		"uuid":  c.Uuid,
	}

	r, err := gorequests.Post(SESSION_ENDPOINT, data, nil, -1)
	if err != nil {
		log.Println(err)
		return
	}

	err = CheckResponse(&r, false)
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
			"token": CROCODOC_API_TOKEN,
			"uuid":  c.Uuid,
		})
	if err != nil {
		log.Println(err)
		return
	}

	r, err := gorequests.Get(GETTEXT_ENDPOINT, data, -1)
	if err != nil {
		log.Println(err)
		return
	}
	err = CheckResponse(&r, false)
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

func (c *CrocoDoc) Download(asPdf bool, filename string, withAnnotations bool, filterUserAnnotations string) (err error) {
	if c.Uuid == "" {
		err = errors.New("Cannot call Download: No UUID is set on the CrocoDoc.")
		return
	}

	var renameTo string
	if !allowedFilename(filename) {
		renameTo = filename
		filename = c.Uuid
	}

	data, err := gorequests.NewQueryData(
		map[string]string{
			"token":     CROCODOC_API_TOKEN,
			"uuid":      c.Uuid,
			"pdf":       asString(asPdf),
			"filename":  filename,
			"annotated": asString(withAnnotations),
			"filter":    filterUserAnnotations,
		})
	if err != nil {
		log.Println(err)
	}

	r, err := gorequests.Get(DOWNLOAD_ENDPOINT, data, -1)
	log.Println(r.Headers())
	if err != nil {
		log.Println(err)
	}

	err = CheckResponse(&r, false)
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
			"token": CROCODOC_API_TOKEN,
			"uuid":  c.Uuid,
			"size":  size,
		})
	if err != nil {
		log.Println(err)
	}

	r, err := gorequests.Get(THUMBNAIL_ENDPOINT, data, -1)
	log.Println(r.Headers())
	if err != nil {
		log.Println(err)
	}

	err = CheckResponse(&r, false)
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
