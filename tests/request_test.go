package crocodoc

import (
	crocodoc "github.com/theplant/crocodoc-go"
	"testing"
	"time"
)

var CDocD, CDocX, CDocP *crocodoc.CrocoDoc // doc, xls, ppt

func TestSetToken(t *testing.T) {
	crocodoc.SetToken("YOUR_TOKEN_HERE")
}

func TestUpload(t *testing.T) {
	// doc
	d, err := crocodoc.UploadFile("testdoc.docx")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	CDocD = d
	t.Log(CDocD)
	// time.Sleep(10 * time.Second)
	// ppt
	d, err = crocodoc.UploadFile("testppt.pptx")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	CDocP = d
	t.Log(CDocP)

	// xls
	// time.Sleep(5 * time.Second) // wait a bit - test accounts are rate limited to 2 cuncurrent conversions
	d, err = crocodoc.UploadFile("testxls.xlsx")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	CDocX = d
	t.Log(CDocX)
}

func TestStatus(t *testing.T) {
	// doc
	t.Log("Checking Doc Status:", CDocD)
	statusResponse, err := crocodoc.GetStatusesForIds([]string{CDocD.Uuid})
	s := statusResponse[0]
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Doc Status:", s)

	if CDocD.Status == crocodoc.QUEUED || CDocD.Status == crocodoc.PROCESSING {
		time.Sleep(10 * time.Second)
		statusResponse, err := crocodoc.GetStatusesForIds([]string{CDocD.Uuid})
		s := statusResponse[0]
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		t.Log("Doc Status:", s)
	}

	// ppt
	t.Log("Checking Doc Status:", CDocP)
	// s, err = CDocP.GetStatus()
	statusResponse, err = crocodoc.GetStatusesForIds([]string{CDocP.Uuid})
	s = statusResponse[0]
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Doc Status:", s)

	if CDocP.Status == crocodoc.QUEUED || CDocP.Status == crocodoc.PROCESSING {
		time.Sleep(10 * time.Second)
		// s, err = CDocP.GetStatus()
		statusResponse, err := crocodoc.GetStatusesForIds([]string{CDocP.Uuid})
		s := statusResponse[0]
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		t.Log("Doc Status:", s)
	}

	// xls
	t.Log("Checking Doc Status:", CDocX)
	// s, err = CDocX.GetStatus()
	statusResponse, err = crocodoc.GetStatusesForIds([]string{CDocX.Uuid})
	s = statusResponse[0]
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Doc Status:", s)

	if CDocX.Status == crocodoc.QUEUED || CDocX.Status == crocodoc.PROCESSING {
		time.Sleep(10 * time.Second)
		// s, err = CDocX.GetStatus()
		statusResponse, err := crocodoc.GetStatusesForIds([]string{CDocX.Uuid})
		s := statusResponse[0]
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		t.Log("Doc Status:", s)
	}
}

func TestGetText(t *testing.T) {
	// doc
	err := CDocD.GetText()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if len(CDocD.ExtractedText) > 100 {
		t.Log(CDocD.ExtractedText[:100])
	} else {
		t.Log(CDocD.ExtractedText)
	}

	// ppt
	err = CDocP.GetText()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if len(CDocP.ExtractedText) > 100 {
		t.Log(CDocP.ExtractedText[:100])
	} else {
		t.Log(CDocP.ExtractedText)
	}

	// xls
	err = CDocX.GetText()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if len(CDocX.ExtractedText) > 100 {
		t.Log(CDocX.ExtractedText[:100])
	} else {
		t.Log(CDocX.ExtractedText)
	}
}

func TestCreateSession(t *testing.T) {
	// doc
	err := CDocD.CreateSession()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(CDocD)

	// ppt
	err = CDocP.CreateSession()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(CDocP)

	// xls
	err = CDocX.CreateSession()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(CDocX)
	t.Log("waiting for 60 seconds... test the session ids...")
	time.Sleep(5 * time.Second)
}

func TestDownload(t *testing.T) {
	// doc
	err := CDocD.Download(false, "foobar.doc", false, "none")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(CDocD)

	// ppt
	err = CDocP.Download(false, "フーバー.ppt", false, "none")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(CDocP)

	// xls
	err = CDocX.Download(false, "foobar.xls", false, "none")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(CDocX)
}

func TestThumbnail(t *testing.T) {
	// doc
	err := CDocD.Thumbnail("", "")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(CDocD)

	// ppt
	err = CDocP.Thumbnail("300x300", "")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(CDocP)

	// xls
	err = CDocX.Thumbnail("", "foobar.png")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(CDocX)
}

func TestDelete(t *testing.T) {
	time.Sleep(10 * time.Second)
	// doc
	del, err := CDocD.Delete()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Deleted?", del)
	t.Log(CDocD)

	// ppt
	del, err = CDocP.Delete()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Deleted?", del)
	t.Log(CDocP)

	// xls
	del, err = CDocX.Delete()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Deleted?", del)
	t.Log(CDocX)
}
