package crocodoc

const (
	// your token here
	// CROCODOC_API_TOKEN = "YOUR_TOKEN_HERE"

	// default location for file downloads
	DEFAULT_FILE_PATH      = "/tmp/"
	DEFAULT_THUMBNAIL_SIZE = "100x100"

	// enpoints
	UPLOAD_ENDPOINT    = "https://crocodoc.com/api/v2/document/upload"
	STATUS_ENDPOINT    = "https://crocodoc.com/api/v2/document/status"
	SESSION_ENDPOINT   = "https://crocodoc.com/api/v2/session/create"
	DELETE_ENDPOINT    = "https://crocodoc.com/api/v2/document/delete"
	DOWNLOAD_ENDPOINT  = "https://crocodoc.com/api/v2/download/document"
	THUMBNAIL_ENDPOINT = "https://crocodoc.com/api/v2/download/thumbnail"
	GETTEXT_ENDPOINT   = "https://crocodoc.com/api/v2/download/text"
	// VIEW_ENDPOINT      = "https://crocodoc.com/view"

	// doc statuses
	QUEUED     = 0 // document conversion has not yet begun
	PROCESSING = 1 // document conversion is in process
	DONE       = 2 // the document was successfully converted
	ERROR      = 3 // an error has occurred during document conversion
	DELETED    = 4

	/* the actual lifetime as stated by crocodoc is 60 minutes. we store the
	creation time on a session and will not rerequest a session within that
	time. to avoid delay related timouts we use 58 minutes instead.
	*/
	SESSION_LIFETIME_MINUTES = 58
)
