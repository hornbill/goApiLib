package apiLib

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var setParamTests = []struct {
	tagName  string
	varValue string
	out      string
}{
	{"User", "admin", ""},
	{"Password", "12345678", ""},
	{"User_name", "blank", ""},
	{"", "blank", "Name Must contain at least one letter or number"},
	{"!\"£$", "blank", "Name invalid only Numbers and letters can be used"},
	{"User2", "!c\"dw", ""},
	{"User3", "<![CDATA[!\"£$%^&*]]>", ""},
	{"co1character", "", ""},
	{"h_site_name", "", ""},
}

func TestSetParam(t *testing.T) {

	for _, tt := range setParamTests {
		conn := NewXmlmcInstance("https://devapi.hornbill.com/test/")
		err := conn.SetParam(tt.tagName, "blank")
		if err != nil {
			if tt.out != err.Error() {
				t.Errorf("for: %s got: %s but want: %s\n", tt.tagName, err, tt.out)
			}
		} else {
			if tt.out != "" {
				t.Errorf("for: %s got: %s but want: %s\n", tt.tagName, err, tt.out)
			}

		}
	}
}

func TestJSONResponse(t *testing.T) {
	conn := NewXmlmcInstance("https://devapi.hornbill.com/test/")
	if conn.jsonresp != false {
		t.Errorf("conn.jsonresp should be false but is %t\n", conn.jsonresp)
	}
	// set to true and test it is
	conn.SetJSONResponse(true)
	if conn.jsonresp != true {
		t.Errorf("conn.jsonresp should be true but is %t\n", conn.jsonresp)
	}

}
func SetAPIKey(t *testing.T) {
	conn := NewXmlmcInstance("https://devapi.hornbill.com/test/")
	// set to true and test it is
	conn.SetAPIKey("testing1234567")
	if conn.apiKey != "testing1234567" {
		t.Errorf("conn.jsonresp should be true but is %t\n", conn.jsonresp)
	}

}

func TestSessionID(t *testing.T) {

	conn := NewXmlmcInstance("https://devapi.hornbill.com/test/")
	if conn.GetSessionID() != "" {
		t.Errorf("conn.GetSessionID() should be blank")
	}
	// Set a fake sessionid
	conn.SetSessionID("ESPSESSION=1278932njfkhi832nfkw9ur8932rk3r932")
	if conn.GetSessionID() != "ESPSESSION=1278932njfkhi832nfkw9ur8932rk3r932" {
		t.Errorf("conn.GetSessionID() should be ESPSESSION=1278932njfkhi832nfkw9ur8932rk3r932")
	}
}

var setElementTests = []struct {
	tagName string
	out     string
}{
	{"User", ""},
	{"Password", ""},
	{"User_name", ""},
	{"", "Element must have at least one letter or number"},
	{"!\"£$", "Element invalid only Numbers and letters can be used"},
	{"User2", ""},
}

func TestOpenElements(t *testing.T) {

	for _, tt := range setElementTests {
		conn := NewXmlmcInstance("https://devapi.hornbill.com/test/")
		err := conn.OpenElement(tt.tagName)
		if err != nil {
			if tt.out != err.Error() {
				t.Errorf("for: %s got: %s but want: %s\n", tt.tagName, err, tt.out)
			}
		} else {
			if tt.out != "" {
				t.Errorf("for: %s got: %s but want: %s\n", tt.tagName, err, tt.out)
			}
		}
	}
}

func TestCloseElements(t *testing.T) {

	for _, tt := range setElementTests {
		conn := NewXmlmcInstance("https://devapi.hornbill.com/test/")
		err := conn.CloseElement(tt.tagName)
		if err != nil {
			if tt.out != err.Error() {
				t.Errorf("for: %s got: %s but want: %s\n", tt.tagName, err, tt.out)
			}
		} else {
			if tt.out != "" {
				t.Errorf("for: %s got: %s but want: %s\n", tt.tagName, err, tt.out)
			}
		}
	}
}

func TestTimeout(t *testing.T) {
	conn := NewXmlmcInstance("https://devapi.hornbill.com/test/")
	if conn.timeout != 30 {
		t.Errorf("Was expecting timeout defualt of 30")
	}
	conn.SetTimeout(40)
	if conn.timeout != 40 {
		t.Errorf("Was expecting timeout of 40 but got%d\n", conn.timeout)
	}
}

func TestGetClearParams(t *testing.T) {
	conn := NewXmlmcInstance("https://devapi.hornbill.com/test/")
	_ = conn.SetParam("stage", "1")
	if conn.GetParam() != "<params><stage>1</stage></params>" {
		t.Errorf("Was expecting %s but got %s\n", `this`, conn.GetParam())
	}
	conn.ClearParam()
	if conn.GetParam() != "<params></params>" {
		t.Errorf("was expecting empty but got %s\n", conn.GetParam())
	}
}
func TestGetZoneInfo(t *testing.T) {

	conn := NewXmlmcInstance("hornbill")
	if conn.server != "https://betaapi.hornbill.com/hornbill/" {
		t.Errorf("Was expecting https://betaapi.hornbill.com/hornbill/ but got %s\n", conn.server)
	}

	conn = NewXmlmcInstance("hTTps://betaapi.hornbill.com/hornbill/")
	if conn.server != "hTTps://betaapi.hornbill.com/hornbill/" {
		t.Errorf("Was expecting hTTps://betaapi.hornbill.com/hornbill/ but got %s\n", conn.server)
	}

	conn = NewXmlmcInstance("NoInstanceNameHERE")
	if conn.server != "" {
		t.Errorf("Was expecting empty but got %s\n", conn.server)
	}

}

var invokeTests = []struct {
	status    int
	jsonre    bool
	useParams bool
	body      string
}{
	{200, false, true, `<?xml version="1.0" encoding="utf-8" ?><methodCallResult status="ok"><params><stageName>XMLMC Session Bind</stageName><nextStage>4</nextStage></params></methodCallResult>`},
	{200, true, true, `{ "@status": true, "params": { "stageName": "XMLMC Session Bind", "nextStage": 4 } }`},
	{500, false, true, `<?xml version="1.0" encoding="utf-8" ?><methodCallResult status="ok"><params><stageName>XMLMC Session Bind</stageName><nextStage>4</nextStage></params></methodCallResult>`},
	{200, false, false, `<?xml version="1.0" encoding="utf-8" ?><methodCallResult status="ok"><params><stageName>XMLMC Session Bind</stageName><nextStage>4</nextStage></params></methodCallResult>`},
}

func TestInvokeXML(t *testing.T) {

	for _, tt := range invokeTests {

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Set-Cookie", "ESPSESSION=jihsuihfndisfdsfjj332njr32")
			w.WriteHeader(tt.status)
			fmt.Fprintf(w, tt.body)
		}))

		defer ts.Close()
		conn := NewXmlmcInstance(ts.URL)
		if tt.useParams == true {
			_ = conn.SetParam("stage", "1")
		}
		conn.SetJSONResponse(tt.jsonre)
		body, err := conn.Invoke("system", "pingCheck")
		if err != nil {
			t.Errorf(err.Error())
		}
		if body != tt.body {
			t.Errorf("Invalid response from invoke:%s\n", body)
		}
		if conn.GetSessionID() != "ESPSESSION=jihsuihfndisfdsfjj332njr32" {
			t.Errorf("Was expecting sessionid as ESPSESSION=jihsuihfndisfdsfjj332njr32 but got %s\n", conn.GetSessionID())
		}
		if conn.GetStatusCode() != tt.status {
			t.Errorf("Was expecting a %d http status code but got %d", tt.status, conn.GetStatusCode())
		}

	}
}
