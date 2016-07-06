package apiLib

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	reg = regexp.MustCompile("^[a-zA-Z0-9_]*$")
)

// XmlmcInstStruct is the struct that contains all the data for a NewXmlmcInstance
type XmlmcInstStruct struct {
	server     string
	paramsxml  string
	statuscode int
	timeout    int
	sessionID  string
	apiKey     string
	jsonresp   bool
}

// ZoneInfoStrut is used to contain the instance zone info data
type ZoneInfoStrut struct {
	Zoneinfo struct {
		Name     string `json:"name"`
		Zone     string `json:"zone"`
		Message  string `json:"message"`
		Endpoint string `json:"endpoint"`
	} `json:"zoneinfo"`
}

// NewXmlmcInstance creates a new xmlc instance. You must pass in the url you are trying to connect to
// and a new instance is returned.
// conn := esp_xmlmc.NewXmlmcInstance("https://eurapi.hornbill.com/test/xmlmc/")
func NewXmlmcInstance(servername string) *XmlmcInstStruct {
	ndb := new(XmlmcInstStruct)
	//-- TK Add Support for passing in instance name
	matchedURL, err := regexp.MatchString(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`, servername)
	//-- Catch Error
	if err != nil {
		log.Fatal("Unable to Parse server Name")
	}
	//-- If URL then just use it
	if matchedURL {
		ndb.server = servername
	} else {
		//-- Else look it up
		ndb.server = GetEndPointFromName(servername)
	}
	ndb.server = servername
	ndb.timeout = 30
	ndb.jsonresp = false
	return ndb
}

// GetEndPointFromName takes an instanceID anf returns a engpoint URL
// looks up json config from https://files.hornbill.com/instances/instanceID/zoneinfo
// serverEndpoint := GetEndPointFromName(servername)
func GetEndPointFromName(instanceID string) string {
	instanceEndpoint := ""

	//-- Get JSON Config
	response, err := http.Get("https://files.hornbill.com/instances/" + instanceID + "/zoneinfo")
	if err != nil || response.StatusCode != 200 {

		//-- If we fail fall over to using files.hornbill.co
		response, err = http.Get("https://files.hornbill.co/instances/" + instanceID + "/zoneinfo")

		//-- If we still have an error then return out
		if err != nil {
			log.Println("Error Loading Zone Info File: " + fmt.Sprintf("%v", err))
			return ""
		}
	}
	//-- Close Connection
	defer response.Body.Close()

	//-- New Decoder
	decoder := json.NewDecoder(response.Body)

	//-- New Var based on ZoneInfoStrut
	zoneInfo := ZoneInfoStrut{}

	//-- Decode JSON
	errDECODE := decoder.Decode(&zoneInfo)

	//-- Error Checking
	if errDECODE != nil {
		log.Println("Error Decoding Zone Info File: " + fmt.Sprintf("%v", errDECODE))
		return ""
	}

	//-- Get Endpoint from URL
	instanceEndpoint = zoneInfo.Zoneinfo.Endpoint
	return instanceEndpoint
}

// SetParam Sets the paramters in an already instantiated NewXmlmcInstance connection.
// returns an errors if this is unsuccesful
// err := conn.SetParam("userId", "admin")
func (xmlmc *XmlmcInstStruct) SetParam(strName string, varValue string) error {
	//Make sure the tag is not empty
	if len(strName) == 0 {
		return errors.New("Name Must contain at least one letter or number")
	}
	//Make sure the tag is only letter ans number so it will create valid XML
	if !reg.MatchString(strName) {
		return errors.New("Name invalid only Numbers and letters can be used")
	}
	//Make sure the ivalues are valid for xml
	cleaned, err := xmlEncodeString(varValue)
	if err != nil {
		return errors.New("Could not clean the varValue input")
	}
	xmlmc.paramsxml = xmlmc.paramsxml + "<" + strName + ">" + cleaned + "</" + strName + ">"
	return nil
}

// Invoke is the call that performs the xml call.
// You pass it the servince name and the methodname as strings.
// It returns the body of the response as a string and an error which should be checked
// result, err := conn.Invoke("session", "userLogon")
func (xmlmc *XmlmcInstStruct) Invoke(servicename string, methodname string) (string, error) {

	xmlmclocal := "<methodCall service=\"" + servicename + "\" method=\"" + methodname + "\">"
	if len(xmlmc.paramsxml) == 0 {
		xmlmclocal = xmlmclocal + "</methodCall>"
	} else {
		xmlmclocal = xmlmclocal + "<params>" + xmlmc.paramsxml
		xmlmclocal = xmlmclocal + "</params>" + "</methodCall>"
	}

	strURL := xmlmc.server + "/" + servicename + "/"

	var xmlmcstr = []byte(xmlmclocal)

	req, err := http.NewRequest("POST", strURL, bytes.NewBuffer(xmlmcstr))

	if err != nil {
		return "", errors.New("Unable to create http request in esp_xmlmc.go")
	}

	req.Header.Set("Content-Type", "text/xmlmc")
	if xmlmc.apiKey != "" {
		req.Header.Add("Authorization", "ESP-APIKEY "+xmlmc.apiKey)
	}
	req.Header.Add("Cookie", xmlmc.sessionID)
	if xmlmc.jsonresp == true {
		req.Header.Add("Accept", "Application/json")
	}
	duration := time.Second * time.Duration(xmlmc.timeout)
	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	client := &http.Client{Transport: t, Timeout: duration}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	xmlmc.statuscode = resp.StatusCode

	defer resp.Body.Close()

	//-- Check for HTTP Response
	if resp.StatusCode != 200 {
		errorString := fmt.Sprintf("Invalid HTTP Response: %d", resp.StatusCode)
		err = errors.New(errorString)
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("Cant read the body of the response")
	}
	// If we have a new EspSessionId set it
	SessionIds := strings.Split(resp.Header.Get("Set-Cookie"), ";")
	if SessionIds[0] != "" {
		xmlmc.sessionID = SessionIds[0]
	}

	xmlmc.paramsxml = ""
	return string(body), nil
}

// SetJSONResponse returns the xml response as json.
// Expects a bool of true or false
// conn.SetJsonResponse(true)
func (xmlmc *XmlmcInstStruct) SetJSONResponse(b bool) {
	xmlmc.jsonresp = b
}

// GetSessionID returins the current set ESPsessionID for this XmlmcInstance
// sessID := conn.GetSessionID()
func (xmlmc *XmlmcInstStruct) GetSessionID() string {
	return xmlmc.sessionID
}

// SetAPIKey sets the current APIKey for this XmlmcInstance it expects a a string to be passed
// conn.SetAPIKey()
func (xmlmc *XmlmcInstStruct) SetAPIKey(s string) {
	xmlmc.apiKey = s
}

// SetSessionID sets the current ESPsessionID for this XmlmcInstance it expects a a string to be passed
// conn.SetSessionID()
func (xmlmc *XmlmcInstStruct) SetSessionID(s string) {
	xmlmc.sessionID = s
}

// GetStatusCode returns the http status code for the last invoked xmlmc call.
// returns an integer
// status := conn.GetStatusCode()
func (xmlmc *XmlmcInstStruct) GetStatusCode() int {
	return xmlmc.statuscode
}

// OpenElement is called to create complex xmlmc requests.
// The name of the element to create should be passed in
// It will return an error if the element name is invalid
// You should always have a matching close tag to match this
// err := conn.OpenElement("UserID")
func (xmlmc *XmlmcInstStruct) OpenElement(elementname string) error {
	//Make sure the element is not empty
	if len(elementname) == 0 {
		return errors.New("Element must have at least one letter or number")
	}
	if !reg.MatchString(elementname) {
		return errors.New("Element invalid only Numbers and letters can be used")
	}
	xmlmc.paramsxml = xmlmc.paramsxml + "<" + elementname + ">"
	return nil
}

// CloseElement is called to close a previously opened OpenElement
// The name of the element to create should be passed in
// It will return an error if the element name is invalid
// err := conn.CloseElement("UserID")
func (xmlmc *XmlmcInstStruct) CloseElement(elementname string) error {
	//Make sure the element is not empty
	if len(elementname) == 0 {
		return errors.New("Element must have at least one letter or number")
	}
	if !reg.MatchString(elementname) {
		return errors.New("Element invalid only Numbers and letters can be used")
	}
	xmlmc.paramsxml = xmlmc.paramsxml + "</" + elementname + ">"
	return nil
}

// SetTimeout allows you to set a maximum timeout for the http request in seconds.
// It defaults to 0 which means no timeout
// This should probably be set to 30 seconds for most requests and should be set before Invoke is called
// conn.SetTimeout = 30
func (xmlmc *XmlmcInstStruct) SetTimeout(timeout int) {
	xmlmc.timeout = timeout
}

func xmlEncodeString(strValue string) (string, error) {

	buf := new(bytes.Buffer)
	err := xml.EscapeText(buf, []byte(strValue))
	if err != nil {
		return "", errors.New(err.Error())
	}
	return buf.String(), nil
}

// GetParam Allows you to get the xml you would be sending to the server.
// It returns a string of the xml
// xmlmc := conn.GetParam()
func (xmlmc *XmlmcInstStruct) GetParam() string {

	return "<params>" + xmlmc.paramsxml + "</params>"
}

// ClearParam Allows you to blank any parms you have already set on a XmlmcInstance
// conn.ClearParam()
func (xmlmc *XmlmcInstStruct) ClearParam() {
	xmlmc.paramsxml = ""
}
