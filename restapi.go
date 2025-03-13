//
//
// restapi
//
// restapi simplifies calling restapis.  It encapsulates a bunch of
// functionality to simplify calls
//
// Also added the XML logic to turn a returned XML file into json
// so we can reuse the same logic in our other apps
//
//

package restapi

import (
	"fmt"
	"bytes"
	"os"
	"net/http"
	"io/ioutil"
        "encoding/json"
        "crypto/tls"
        "crypto/x509"
        "strings"
        "time"
        "github.com/basgys/goxml2json"   
        "github.com/seldonsmule/logmsg"
        "github.com/twpayne/go-jsonstruct"

)

type HttpMethod int


// Type of API methods

const (
        Get HttpMethod = 1 + iota
        Post
        Put
        Delete
)


//
// our internal control structure
//

type Restapi struct {

  sAccessToken         string
  sUrl                 string
  sName                string
  Method               HttpMethod
  sMethodString        string

  bRequiresAccessToken       bool
  bRequiresApiKey            bool
  bInnerMap                  bool
  bHasPostJson               bool
  bInnerMapArray             bool
  sInnerMapName              string
  bDebug                     bool
  bXML                       bool
  bXMLDontParseResponse      bool

  bJsonOnly                  bool // if true, we don't want the extra map help

  sCertFile                  string
  bUseCertFile               bool
  pcaCertPool                  *x509.CertPool

  sJsonStr string

  nLastStatusCode int

  RawData interface{}  // used to contain the raw response msg mody
  BodyString string
  BodyBytes []byte

  mResponseMapData map[string]interface{}
  mInnerMapData map[string]interface{}

  amInnerMapArray []interface{}
  sInnerMapArrayCountName  string
  iInnerMapArrayCount int

}

//
// func NewPost(name string, url string) *Restapi
//
// Create a new restapi object for sending post
//
// name - name of the post
// url - URL to execute against
//
//

func NewPost(name string, url string) *Restapi{
  return(New(Post, name, url))
}

//
// func (pRA *Restapi) UseCert(certfile string) bool
//
// stores certificate from file for use
//
// certfil - file with certificate to be trusted
//
//

func (pRA *Restapi) UseCert(certfile string) bool {

  pRA.bUseCertFile = true
  pRA.sCertFile = certfile

/*
  caCert, err := ioutil.ReadFile("powerwall.cer")

  caCertPool := x509.NewCertPool()
  caCertPool.AppendCertsFromPEM(caCert)

  if (err != nil){
    msg := fmt.Sprintf("Error opening certificate file [%s] %s\n", certfile, err)
    logmsg.Print(logmsg.Error, msg)
    return false
  }
*/

  return true
}

//
// FetchTLSCert queries the gateway and returns a copy of the TLS certificate
// it is currently presenting for connections.  This is useful for saving and
// later using with `SetTLSCert` to validate future connections.
//
// Credits - this code came from https://github.com/foogod/go-powerwall/blob/main/client.go
// 
// Might have to rethink a bunch based on his more elegant use of go
//
//

func (pRA *Restapi) FetchTLSCert(url string) (*x509.Certificate, bool) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", url+":443", tlsConfig)
	if err != nil {
          msg := fmt.Sprintf("Error getting certificat", err)
          logmsg.Print(logmsg.Error, msg)
		return nil, false
	}
///fmt.Println(tlsConfig)
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
///fmt.Println(certs)

	return certs[0], true
}



//
// func NewGet(name string, url string) *Restapi
//
// Create a new restapi object for sending GET
//
// name - name of the get
// url - URL to execute against
//
//


func NewGet(name string, url string) *Restapi{
  return(New(Get, name, url))
}

//
// func NewGetXML(name string, url string, parseresponse bool) *Restapi
//
// Create a new restapi object for calling a URL tha responds with XML
//
// name - name of the cmd - more of a reference thing for logging
// url - URL to execute against
// parseresponse - true/false. If true, will attempt load into json
//
//

func NewGetXML(name string, url string, parseresponse bool) *Restapi{

  r := New(Get, name, url)
  r.bXML = true
  r.bXMLDontParseResponse = parseresponse

  return r
}

//
// func NewPut(name string, url string) *Restapi 
//
// Create a new restapi object for sending PUT
//
// name - name of the put
// url - URL to execute against
//
//

func NewPut(name string, url string) *Restapi{
  return(New(Put, name, url))
}

//
// func NewDelete(name string, url string) *Restapi
//
// Create a new restapi object for sending DELETE
//
// name - name of the delete
// url - URL to execute against
//
//

func NewDelete(name string, url string) *Restapi{
  return(New(Delete, name, url))
}

//
// func New(method HttpMethod, name string, url string) *Restapi
//
// Create a new restapi object
//
// method - Type of http method (get, put, delete, etc)
// name - name of the get
// url - URL to execute against
//
//

func New(method HttpMethod, name string, url string) *Restapi{


  logmsg.Print(logmsg.Info, "In Restapi New")

  r := new(Restapi)

  r.bRequiresAccessToken = false
  r.bRequiresApiKey = false


  r.setUrl(url)

  r.bUseCertFile = false
  r.bJsonOnly = false

  r.sName = name
  r.setMethod(method)

  r.bInnerMap = false
  r.bHasPostJson = false
  r.bInnerMapArray = false
  r.DebugOff()

  return r
}

//
// func (pRA *Restapi) GetLastStatusCode() int
//
// Get the last status code
//

func (pRA *Restapi) GetLastStatusCode() int{
  return pRA.nLastStatusCode
}

//
// func (pRA *Restapi) GetArrayValueString(index int, key string) string{
//
// Return string value of an array index.  
//
// index - index into array
// key   - key string being looked for in the map from the index
//
//

func (pRA *Restapi) GetArrayValueString(index int, key string) string{
  return(CastString(pRA.GetArrayValue(index, key)))
}

//
// func (pRA *Restapi) GetArrayValueInt(index int, key string) int
//
// Return integer value of an array index
//
// index - index into array
// key   - key string being looked for in the map from the index
//
//

func (pRA *Restapi) GetArrayValueInt(index int, key string) int{
  return(CastFloatToInt(pRA.GetArrayValue(index, key)))
}

//
// func (pRA *Restapi) GetArrayValueInt64(index int, key string) uint64
//
// Return integer64 value of an array index
//
// index - index into array
// key   - key string being looked for in the map from the index
//
//

func (pRA *Restapi) GetArrayValueInt64(index int, key string) uint64{
  return(CastFloatToInt64(pRA.GetArrayValue(index, key)))
}

//
// func (pRA *Restapi) GetArrayValue(index int, key string) interface{}
//
// Return value of an array index
//
// index - index into array
// key   - key string being looked for in the map from the index
//
//

func (pRA *Restapi) GetArrayValue(index int, key string) interface{}{

  if(!pRA.bInnerMapArray){
    logmsg.Print(logmsg.Error, "No inner array set")
    return nil
  }

  if(index >= pRA.iInnerMapArrayCount){
    logmsg.Print(logmsg.Error, "Index outside of array Range")
  }

  tmpmap := CastMap(pRA.amInnerMapArray[index])

  return(tmpmap[key])

}

//
// func (pRA *Restapi) GetValue(index string) interface{}
//
// Gets the value of a map
//
// index - map string name
//

func (pRA *Restapi) GetValue(index string) interface{}{

  // note interface{} is similar to void in C (in my mind)
  // you will have to type cast the results to use

  if(pRA.bInnerMap){
    return(pRA.mInnerMapData[index])
  }else{
    return(pRA.mResponseMapData[index])
  }

}

//
//
// func (pRA *Restapi) GetValueString(index string) string
//
// calls getvalue and cast it as a string
//
// index - string index being looked for
//

func (pRA *Restapi) GetValueString(index string) string{

  return(CastString(pRA.GetValue(index)))

}

//
//
// func (pRA *Restapi) GetValueInt(index string) int
//
// calls getvalue and cast it as an int
//
// index - string index being looked for
//

func (pRA *Restapi) GetValueInt(index string) int{

  return(CastFloatToInt(pRA.GetValue(index)))

}

//
// CastArray - Sorry I like C's terminology so built a quick
//             helper function

func CastArray(item interface{}) []interface{} {

  return item.([]interface{})

}

//
// CastFloatToInt - Sorry I like C's terminology so built a quick
//             helper function

func CastFloatToInt(item interface{}) int {

  var f float64
  f = item.(float64)

  return int(f)

}

//
// CastFloatToInt64 - Sorry I like C's terminology so built a quick
//             helper function

func CastFloatToInt64(item interface{}) uint64 {

  var f float64
  f = item.(float64)

  return uint64(f)

}

//
// CastString - Sorry I like C's terminology so built a quick
//             helper function

func CastString(item interface{}) string {
  return item.(string)
}


//
// CastMap - Sorry I like C's terminology so built a quick
//             helper function

func CastMap(item interface{}) map[string]interface{} {

  return(item.(map[string]interface{}))

}

//
// func (pRA *Restapi) Dump()
//
// For Diagnostics - dumps out the contents
//

func (pRA *Restapi) Dump(){

  fmt.Println("Dump:", pRA.sName)
  fmt.Println("Url:", pRA.sUrl)
  fmt.Println("Method:", pRA.Method)
  fmt.Println("MethodString:", pRA.sMethodString)
  fmt.Println("AccessToken:", pRA.sAccessToken)
  fmt.Println("UseCert:", pRA.bUseCertFile)
  fmt.Println("JsonStr:", pRA.sJsonStr)

  if(pRA.bUseCertFile){
    fmt.Println("sCertFile:",pRA.sCertFile)
  }
  
  if(pRA.bInnerMap){
    fmt.Println("sInnerMapName:",pRA.sInnerMapName)
  }

  fmt.Println("ResponseMapData:")
  for name, value := range pRA.mResponseMapData {
    fmt.Println(name, "=", value)
  }

  if( pRA.bInnerMap ){
    fmt.Println("InnerMapData:")
    for k, v := range pRA.mInnerMapData {
      fmt.Println(k, "=", v)

    } // end for loop
  }

  if( pRA.bInnerMapArray ){
    fmt.Println("InnerMapArray")
    fmt.Println("InnerMapArrayCount:", pRA.iInnerMapArrayCount)
    
    for i:=0 ; i < pRA.iInnerMapArrayCount; i++ {
      //fmt.Println("Index:", i)
      //fmt.Println(pRA.amInnerMapArray[i])
      tmpmap := CastMap(pRA.amInnerMapArray[i])
      for k, v := range tmpmap {
        fmt.Println("Index 0:", k, "=", v)

      } // end for loop
    }
  }

/*
  switch vv := v.(type) {

    case string:
      fmt.Println(k, "is string", vv)

    case bool:
      fmt.Println(k, "is bool", vv)

    case int64:
      fmt.Println(k, "is int64", vv)

    case float64:
      fmt.Println(k, "is float64", vv)

    case int:
      fmt.Println(k, "is int", vv)

    case map[string]interface{}:
      fmt.Println(k, "Another Map", vv)

    default:
      fmt.Println(k, "is of a type I don't know how to handle")

  }
*/


}

//
// func (pRA *Restapi) DebugOn()
//
// Turn on debugging.  This will generate extra dumps of data
// to standard out
//

func (pRA *Restapi) DebugOn(){
  pRA.bDebug = true
}

//
// func (pRA *Restapi) DebugOff()
//
// Turn off debugging
//

func (pRA *Restapi) DebugOff(){
  pRA.bDebug = false
}

//
// func (pRA *Restapi) JsonOnly()
//
// Skips tryiing to process map data of the converted response
//

func (pRA *Restapi) JsonOnly(){
  pRA.bJsonOnly = true
}

//
// func (pRA *Restapi) HasInnerMap(name string)
//
// This was designed around the tesla apis that has a very meassured
// way of nesting a map within a map.  You put in the name of the map
// when setting andn restapi will setup a pointer directly to it 
// for easy access
//
// Truthfully - probably a tesla thing only, but that was what
// restapi was originally created to simplify the code for
// 
// name - name of intermap 
//

func (pRA *Restapi) HasInnerMap(name string){
  pRA.bInnerMap = true
  pRA.sInnerMapName = name
}

//
// func (pRA *Restapi) HasInnerMapArray(name string, countname string)
//
// Same update as HasInnerMap().  But it is an array of maps
// 
// name - name of intermap 
// countname - Tesla specific - map name containing array count
//

func (pRA *Restapi) HasInnerMapArray(name string, countname string){
  pRA.bInnerMapArray = true
  pRA.sInnerMapName = name
  pRA.sInnerMapArrayCountName = countname
  pRA.iInnerMapArrayCount = 0
}

//
// func (pRA *Restapi) SetApiKey(ApiKey string)
//
// Sets Api Key for Authentication
//

func (pRA *Restapi) SetApiKey(ApiKey string){
  //pRA.sAccessToken = fmt.Sprintf("X-API_KEY %s", ApiKey)
  pRA.sAccessToken = ApiKey
  pRA.bRequiresApiKey = true
}

//
// func (pRA *Restapi) SetBearerAccessToken(AccessToken string)
//
// Sets Token for Bearer Authentication
//

func (pRA *Restapi) SetBearerAccessToken(AccessToken string){
  pRA.sAccessToken = fmt.Sprintf("Bearer %s", AccessToken)
  pRA.bRequiresAccessToken = true
}

//
//func (pRA *Restapi) SetBasicAccessToken(AccessToken string){
//
// Sets Token for Basic Authentication
//

func (pRA *Restapi) SetBasicAccessToken(AccessToken string){
  pRA.sAccessToken = fmt.Sprintf("Basic %s", AccessToken)
  pRA.bRequiresAccessToken = true
}

//
//
// func (pRA *Restapi) setUrl(Url string){
//
// store the URL string to call
// 
// Url - Url of API
//

func (pRA *Restapi) setUrl(Url string){
  pRA.sUrl = Url
}

func (pRA *Restapi) GetUrl() string{
  return(pRA.sUrl)
}

func (pRA *Restapi) GetName() string{
  return(pRA.sName)
}

//
//
// Sets the method of API
//
// method - Nemonic for API type
//          See HttpMethod struct
// 

func (pRA *Restapi) setMethod(method HttpMethod){

  pRA.Method = method

  switch method {
    case Post:
      pRA.sMethodString = "POST"

    case Get:
      pRA.sMethodString = "GET"

    case Put:
      pRA.sMethodString = "PUT"

    case Delete:
      pRA.sMethodString = "DELETE"

    default:
      pRA.sMethodString = "WHO KNOWS"

  }
}

//
// func TurnOffCertValidation()
//
// Added this for dealing with known self signed certs.  
// Otherwise https call will fail
//

func TurnOffCertValidation(){

  http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

}

//
// func TurnOnCertValidation()
//
// Added this for dealing with known self signed certs.  
// Otherwise https call will fail
//

func TurnOnCertValidation(){

  http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}

}

//
// func (pRA *Restapi) GetResponseBody() string
//
// Returns the string version of the response
//

func (pRA *Restapi) GetResponseBody() string {
  return pRA.BodyString
}

//
// func (pRA *Restapi) SaveResponseBody(filename string, bstdout bool) bool
//
// Used to save off the data to build out json structs
//

func (pRA *Restapi) SaveResponseBody(filename string, structname string, bstdout bool) bool {

  var observedValue *jsonstruct.ObservedValue
  var decoded interface{}

  json_fhd, err := os.Create(filename+".json")

  if(err != nil){
    msg := fmt.Sprintf("SaveResponseBody Create File error[%s]", err)
    logmsg.Print(logmsg.Error, msg)
    return false
  }

  defer json_fhd.Close()

  go_fhd, gerr := os.Create(filename+".go")

  if(gerr != nil){
    msg := fmt.Sprintf("SaveResponseBody Create File error[%s]", gerr)
    logmsg.Print(logmsg.Error, msg)
    return false
  }

  defer go_fhd.Close()

  json.Unmarshal(pRA.BodyBytes, &decoded)

  observedValue = observedValue.Merge(decoded)

  comment := "This file was autogenerated using https://github.com/twpayne/go-jsonstruct\n\n"

  options := []jsonstruct.GeneratorOption{
                jsonstruct.WithOmitEmpty(jsonstruct.OmitEmptyAuto),
                jsonstruct.WithSkipUnparseableProperties(true),
                jsonstruct.WithUseJSONNumber(false),
                jsonstruct.WithGoFormat(true),
		jsonstruct.WithTypeName(structname),
		jsonstruct.WithPackageComment(comment),
  
              }

  goCode, jerr := jsonstruct.NewGenerator(options...).GoCode(observedValue)

  if(jerr != nil) {
    msg := fmt.Sprintf("SaveResponseBody jsonstruct.NewGenerator error[%s]", jerr)
    logmsg.Print(logmsg.Error, msg)
    return false
  }

  
  if(bstdout){
    os.Stdout.Write(goCode)
  }
  
  _, err = go_fhd.Write(goCode)

  if(err != nil){
    msg := fmt.Sprintf("SaveResponseBody Write File err[%s]", err)
    logmsg.Print(logmsg.Error, msg)
    return false
  }
  
  var prettyJSON bytes.Buffer

  pretty_err := json.Indent(&prettyJSON, pRA.BodyBytes, "", "  ")

  if(pretty_err != nil){

    msg := fmt.Sprintf("SaveResponseBody json.Indent err[%s]", err)
    logmsg.Print(logmsg.Error, msg)
    return false
  }

  //_, err2 := json_fhd.WriteString(pRA.GetResponseBody())
  _, err2 := json_fhd.WriteString(prettyJSON.String())
  

  if(err2 != nil){
    msg := fmt.Sprintf("SaveResponseBody Write File err2[%s]", err2)
    logmsg.Print(logmsg.Error, msg)
    return false
  }

  return true
}

//
// func (pRA *Restapi) Send() bool
//
// Sends the API request
//

func (pRA *Restapi) Send() bool {

  var req *http.Request
  var tran *http.Transport

  tran = nil

  if(len(pRA.sUrl) == 0){
    msg := fmt.Sprintf("Send(%s): Url not set", pRA.sName)
    logmsg.Print(logmsg.Error, msg)
    if(pRA.bDebug){
      fmt.Println(msg)
    }
    return false
  }

  if(pRA.bDebug){
    fmt.Println("URL:",pRA.sUrl)
  }

  if(!pRA.bHasPostJson){
    req, _ = http.NewRequest(pRA.sMethodString, pRA.sUrl, nil)
  }else{
    req, _ = http.NewRequest(pRA.sMethodString, pRA.sUrl, 
                              bytes.NewBufferString(pRA.sJsonStr))
  }

  if(pRA.bRequiresAccessToken){
    req.Header.Add("Authorization", pRA.sAccessToken)
  }

  if(pRA.bRequiresApiKey){
    req.Header.Add("x-api-key", pRA.sAccessToken)
  }

//  req.Header.Add("Accept", "*/*")

  req.Header.Add("cache-control", "no-cache")
  req.Header.Add("Content-Type", "application/json")


//fmt.Println(req)

  

  if(pRA.bUseCertFile){
    caCert, err := ioutil.ReadFile(pRA.sCertFile)

    if(err != nil){
      msg := fmt.Sprintf("Error reading cert file[%s] - %s", pRA.sCertFile, err)
      logmsg.Print(logmsg.Error, msg)
      return false 
    }

    pRA.pcaCertPool = x509.NewCertPool()
    pRA.pcaCertPool.AppendCertsFromPEM(caCert)


    tran = &http.Transport{ TLSClientConfig: &tls.Config{ RootCAs: pRA.pcaCertPool } }

  } // end if use a certfile


  if(pRA.bDebug){
    fmt.Println(req)
  }

//  var netClient = &http.Client{Timeout: time.Second * 10, }
  var netClient *http.Client

  if(tran == nil){
    netClient = &http.Client{Timeout: time.Second * 10, }
  }else{
    netClient = &http.Client{Transport: tran, Timeout: time.Second * 10, }
  }

  //res, err := http.DefaultClient.Do(req)
  res, err := netClient.Do(req)

  if(res == nil){
    //logmsg.Print(logmsg.Error, "Error getting to server at URL:", pRA.sUrl)
    logmsg.Print(logmsg.Error, "Error:", err)
    if(pRA.bDebug){
      fmt.Println("Error getting to server at URL:", pRA.sUrl)
    }
    return false
  }

  if(pRA.bDebug){
    fmt.Println("HTTP Response Status:", res.StatusCode, http.StatusText(res.StatusCode))
  }

  pRA.nLastStatusCode = res.StatusCode

  switch res.StatusCode {

    case 200:
    case 201:

    default:
      logmsg.Print(logmsg.Error,"HTTP Response Status:", res.StatusCode, http.StatusText(res.StatusCode))
      logmsg.Print(logmsg.Error,pRA.sUrl)
      return false

  }


  defer res.Body.Close()
  body, _ := ioutil.ReadAll(res.Body)

  pRA.BodyBytes = body
  pRA.BodyString = string(body) // save this off


  if(pRA.bDebug){
    fmt.Println(res)
    fmt.Println(string(body))
  }

//
// added xml logic 9/8/2019 
// using the xml2jon library from github we move the xml into json
// and go back to json work
//
  if(pRA.bXML){
    if(!pRA.bXMLDontParseResponse){
      pRA.RawData = string(body) // need to figure out how to save
      return true
    }

/////fmt.Println(string(body))

    xml := strings.NewReader(string(body))

    ejson, err := xml2json.Convert(xml)
    if err != nil {
  	panic("That's embarrassing...")
    }
/////fmt.Println(ejson.String())

   // reusing the body variable so we can fall through to exising logic
   // pre-xml code added

   body = []byte(ejson.String())

  }

  json.Unmarshal(body, &pRA.RawData)

/////fmt.Println(pRA.RawData)
/////os.Exit(1)

  if(pRA.bDebug){
    fmt.Println(pRA.RawData)
  }

  if(pRA.RawData == nil){

    logmsg.Print(logmsg.Warning,"No data returned")

    return true
  }

  // This test is because some of the processing below still does 
  // not always work.  Allows you to use the tool and work 
  // work with just the json responses if desired and skip
  // the map building.  
  //
  // Note the issue is the json response does not match a pattern
  
  if(pRA.bJsonOnly){

    logmsg.Print(logmsg.Warning,"bJsonOnly set - data not fully processed")
    return true

  }

  pRA.mResponseMapData = CastMap(pRA.RawData)

  if(pRA.bInnerMap){
    if(pRA.bDebug){
      fmt.Println("Looking for innermap:", pRA.sInnerMapName)
    }
    pRA.mInnerMapData = CastMap(pRA.mResponseMapData[pRA.sInnerMapName])
  }else if(pRA.bInnerMapArray){
    tmp1 := pRA.mResponseMapData[pRA.sInnerMapName]
    pRA.amInnerMapArray = CastArray(tmp1)

    var f float64
    f = pRA.mResponseMapData[pRA.sInnerMapArrayCountName].(float64)

    pRA.iInnerMapArrayCount = int(f)

  }


  return true


}

func (pRA *Restapi) SetPostJson(jsonstr string) bool {

  pRA.bHasPostJson = true
  pRA.sJsonStr = jsonstr

//fmt.Println(pRA.sJsonStr)

  return true

}

