crocodoc-go
===========

Wrapper/library for the [crocodoc REST API](https://crocodoc.com/docs/api). This is a first version ...

godoc documentation is [here](http://godoc.org/github.com/theplant/crocodoc-go).

Some examples
-------------

General

```
import (
	crocodoc "crocodoc-go"
    "log"
    "time"
)

var cdoc *crocodoc.CrocoDoc // doc file

func init() {
	crocodoc.SetToken("YOUR_API_TOKEN")
}
```


Upload a doc

```
d, err := crocodoc.UploadFile("testdoc.docx")
if err != nil {
   // do something
}
```

`UploadFile` is actually a wrapper for `Upload(fs io.ReadCloser, filename string)` which you can use to upload a file to crocodoc that comes from any `io.ReadCloser` stream.

We initialized the doc object by uploading it. If you have an id from crocodoc stored somewhere and want to call any of the following functions you could also initialize it just with that id

```
cdoc := &crocodoc.CrocoDoc{
	Uuid: "somecrocodocid",
}
```

Check the status of that doc until its processed (or had a processing error)

```
s, err := cdoc.GetStatus()
if err != nil {
   // do something
}
log.Println("Doc Status:", s)

if cdoc.Status == crocodoc.QUEUED || cdoc.Status == crocodoc.PROCESSING {
	time.Sleep(2 * time.Second)
	s, err = cdoc.GetStatus()
	if err != nil {
        // do something
	}
	log.Println("Doc Status:", s)
}
```

Create a viewing session for the crocodoc HTML5 viewer

```
err := cdoc.CreateSession()
if err != nil {
   // do something
}
log.Println("Session id:", cdoc.SessionId)
```

Extract the text from the file

```
err := cdoc.GetText()
if err != nil {
   // do something
}
log.Println(cdoc.ExtractedText)
```

Download the document

```
err := cdoc.Download(false, "foobar.doc", false, "none")
if err != nil {
   // do something
}
```

Download a thumbnail of the document

```
err := cdoc.Thumbnail("300x300", "foobar.png")
if err != nil {
   // do something
}
```

Delete the document

```
del, err := cdoc.Delete()
if err != nil {
   // do something
}
log.Println("Deleted?", del)
log.Println(cdoc.Status)
```