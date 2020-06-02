package apiLib

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
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
	server      string
	stream      string
	DavEndpoint string
	FileError   error
	paramsxml   string
	statuscode  int
	timeout     int
	count       uint64
	sessionID   string
	apiKey      string
	trace       string
	jsonresp    bool
	userAgent   string
	transport   *http.Transport
}

// ZoneInfoStrut is used to contain the instance zone info data
type ZoneInfoStrut struct {
	Zoneinfo struct {
		Name     string `json:"name"`
		Zone     string `json:"zone"`
		Message  string `json:"message"`
		Endpoint string `json:"endpoint"`
		Stream   string `json:"releaseStream"`
	} `json:"zoneinfo"`
}

// ParamAttribStruct is used to set XML attribues on an XMLMC parameter
type ParamAttribStruct struct {
	Name  string
	Value string
}

// NewXmlmcInstance creates a new xmlc instance. You must pass in the url you are trying to connect to
// and a new instance is returned.
// conn := esp_xmlmc.NewXmlmcInstance("https://eurapi.hornbill.com/test/xmlmc/")
func NewXmlmcInstance(servername string) *XmlmcInstStruct {
	ndb := new(XmlmcInstStruct)
	//-- TK Add Support for passing in instance name
	matchedURL, err := regexp.MatchString(`(?i)(http|https)(?-i):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`, servername)
	//-- Catch Error
	if err != nil {
		log.Fatal("Unable to Parse server Name")
	}
	//-- If URL then just use it
	if matchedURL {
		ndb.server = servername
		ndb.DavEndpoint = strings.Replace(servername, "/xmlmc/", "/dav/", 1)
	} else {
		//-- Else look it up
		serverZoneInfo, ziErr := GetZoneInfo(servername)
		ndb.FileError = ziErr
		if ziErr != nil {
			return ndb
		}
		if serverZoneInfo.Zoneinfo.Endpoint != "" {
			ndb.server = serverZoneInfo.Zoneinfo.Endpoint + "xmlmc/"
			ndb.DavEndpoint = serverZoneInfo.Zoneinfo.Endpoint + "dav/"
		}
		if serverZoneInfo.Zoneinfo.Stream != "" {
			ndb.stream = serverZoneInfo.Zoneinfo.Stream
		}
	}
	ndb.transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	ndb.userAgent = "Go-http-client/1.1"
	ndb.timeout = 30
	ndb.jsonresp = false
	return ndb
}

// GetEndPointFromName takes an instanceID anf returns a endpoint URL
// looks up json config from https://files.hornbill.com/instances/instanceID/zoneinfo
// serverEndpoint := GetEndPointFromName(servername)
func GetEndPointFromName(instanceID string) string {
	if instanceID == "" {
		return ""
	}
	instanceZoneInfo, _ := GetZoneInfo(instanceID)
	return instanceZoneInfo.Zoneinfo.Endpoint
}

// GetZoneInfo takes an instance ID and returns zone information about the instance
// looks up json config from https://files.hornbill.com/instances/instanceID/zoneinfo
// instanceZoneInfo := GetZoneInfo(servername)
func GetZoneInfo(instanceID string) (ZoneInfoStrut, error) {
	//-- New Var based on ZoneInfoStrut
	zoneInfo := ZoneInfoStrut{}
	if instanceID == "" {
		return zoneInfo, errors.New("instanceid not provided")
	}
	//-- Get JSON Config
	response, err := http.Get("https://files.hornbill.com/instances/" + instanceID + "/zoneinfo")
	if err != nil || response.StatusCode != 200 {

		//-- If we fail fall over to using files.hornbill.co
		response, err = http.Get("https://files.hornbill.co/instances/" + instanceID + "/zoneinfo")

		//-- If we still have an error then return out
		if err != nil {
			log.Println("Error Loading Zone Info File: " + err.Error())
			return zoneInfo, err
		}
	}
	//-- Close Connection
	defer response.Body.Close()

	//-- New Decoder
	decoder := json.NewDecoder(response.Body)

	//-- Decode JSON
	errDECODE := decoder.Decode(&zoneInfo)

	//-- Error Checking
	if errDECODE != nil {
		log.Println("Error Decoding Zone Info File:", errDECODE.Error())
		return zoneInfo, errDECODE
	}
	return zoneInfo, nil
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

// SetParamAttr sets a parameter with attributes in an already instantiated NewXmlmcInstance connection.
// returns an errors if this is unsuccesful
// err := conn.SetParamAttr("userId", "admin", yourAttribsArray)
func (xmlmc *XmlmcInstStruct) SetParamAttr(strName string, varValue string, attribs []ParamAttribStruct) error {
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
	xmlmc.paramsxml = xmlmc.paramsxml + "<" + strName
	for _, v := range attribs {
		xmlmc.paramsxml += " " + v.Name + "=\"" + v.Value + "\""
	}
	xmlmc.paramsxml = xmlmc.paramsxml + ">" + cleaned + "</" + strName + ">"
	return nil
}

// InvokeGetResponse is the call that performs the xml call.
// You pass it the servince name and the methodname as strings.
// It returns the body of the response as a string the http response headers and an error which should be checked
// result, err := conn.Invoke("session", "userLogon")
func (xmlmc *XmlmcInstStruct) InvokeGetResponse(servicename string, methodname string) (string, http.Header, error) {

	//-- Add Api Tracing
	tracename := ""
	if xmlmc.trace != "" {
		tracename = "/" + tracename
	}

	xmlmclocal := "<methodCall service=\"" + servicename + "\" method=\"" + methodname + "\" trace=\"goApi" + tracename + "\">"
	if len(xmlmc.paramsxml) == 0 {
		xmlmclocal = xmlmclocal + "</methodCall>"
	} else {
		xmlmclocal = xmlmclocal + "<params>" + xmlmc.paramsxml
		xmlmclocal = xmlmclocal + "</params>" + "</methodCall>"
	}

	strURL := xmlmc.server + "/" + servicename + "/?method=" + methodname

	var xmlmcstr = []byte(xmlmclocal)

	req, err := http.NewRequest("POST", strURL, bytes.NewBuffer(xmlmcstr))
	xmlmc.count++

	if err != nil {
		return "", nil, errors.New("Unable to create http request in esp_xmlmc.go")
	}

	req.Header.Set("Content-Type", "text/xmlmc")
	if xmlmc.apiKey != "" {
		req.Header.Add("Authorization", "ESP-APIKEY "+xmlmc.apiKey)
	}
	req.Header.Set("User-Agent", xmlmc.userAgent)
	req.Header.Add("Cookie", xmlmc.sessionID)
	if xmlmc.jsonresp == true {
		req.Header.Add("Accept", "text/json")
	}
	duration := time.Second * time.Duration(xmlmc.timeout)
	client := &http.Client{Transport: xmlmc.transport, Timeout: duration}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	xmlmc.statuscode = resp.StatusCode

	defer resp.Body.Close()
	//-- Check for HTTP Response
	if resp.StatusCode != 200 {
		errorString := fmt.Sprintf("Invalid HTTP Response: %d", resp.StatusCode)
		err = errors.New(errorString)
		//Drain the body so we can reuse the connection
		io.Copy(ioutil.Discard, resp.Body)
		return "", nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, errors.New("Cant read the body of the response")
	}
	// If we have a new EspSessionId set it
	SessionIds := strings.Split(resp.Header.Get("Set-Cookie"), ";")

	if SessionIds[0] != "" {
		xmlmc.sessionID = SessionIds[0]
	}

	xmlmc.paramsxml = ""
	return string(body), resp.Header, nil
}

// Invoke is the call that performs the xml call.
// You pass it the servince name and the methodname as strings.
// It returns the body of the response as a string and an error which should be checked
// result, err := conn.Invoke("session", "userLogon")
func (xmlmc *XmlmcInstStruct) Invoke(servicename string, methodname string) (string, error) {

	//-- Add Api Tracing
	tracename := ""
	if xmlmc.trace != "" {
		tracename = "/" + tracename
	}

	xmlmclocal := "<methodCall service=\"" + servicename + "\" method=\"" + methodname + "\" trace=\"goApi" + tracename + "\">"
	if len(xmlmc.paramsxml) == 0 {
		xmlmclocal = xmlmclocal + "</methodCall>"
	} else {
		xmlmclocal = xmlmclocal + "<params>" + xmlmc.paramsxml
		xmlmclocal = xmlmclocal + "</params>" + "</methodCall>"
	}

	strURL := xmlmc.server + "/" + servicename + "/?method=" + methodname

	var xmlmcstr = []byte(xmlmclocal)

	req, err := http.NewRequest("POST", strURL, bytes.NewBuffer(xmlmcstr))
	xmlmc.count++

	if err != nil {
		return "", errors.New("Unable to create http request in esp_xmlmc.go")
	}

	req.Header.Set("Content-Type", "text/xmlmc")
	if xmlmc.apiKey != "" {
		req.Header.Add("Authorization", "ESP-APIKEY "+xmlmc.apiKey)
	}
	req.Header.Set("User-Agent", xmlmc.userAgent)
	req.Header.Add("Cookie", xmlmc.sessionID)
	if xmlmc.jsonresp == true {
		req.Header.Add("Accept", "text/json")
	}
	duration := time.Second * time.Duration(xmlmc.timeout)
	client := &http.Client{Transport: xmlmc.transport, Timeout: duration}
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
		//Drain the body so we can reuse the connection
		io.Copy(ioutil.Discard, resp.Body)
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

// SetTrace sets the current Trace to goAPI/STRING for this XmlmcInstance it expects a a string to be passed
// conn.SetTrace()
func (xmlmc *XmlmcInstStruct) SetTrace(s string) {
	xmlmc.trace = s
}

// GetServerURL returns the URL of the tenant used by the xmlmc call
// server := conn.GetServerURL()
func (xmlmc *XmlmcInstStruct) GetServerURL() string {
	return xmlmc.server
}

// GetServerStream returns the stream of the instance
// stream := conn.GetServerStream
func (xmlmc *XmlmcInstStruct) GetServerStream() string {
	return xmlmc.stream
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

// SetUserAgent Sets a new userAgent to be passed in so we can identify who is sending the requests
// conn.SetUserAgent("Ldap import tool")
func (xmlmc *XmlmcInstStruct) SetUserAgent(ua string) {
	xmlmc.userAgent = ua
}

// GetCount Allows you to get the Count of API Calls that have been made.
// It returns a string of the xml
// xmlmc := conn.GetCount()
func (xmlmc *XmlmcInstStruct) GetCount() uint64 {

	return xmlmc.count
}
