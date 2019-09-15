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
	"net/http"
	"io/ioutil"
        "encoding/json"
        "crypto/tls"
        "strings"
        "github.com/basgys/goxml2json"   
        "github.com/seldonsmule/logmsg"

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
  bInnerMap                  bool
  bInnerMapArray             bool
  sInnerMapName              string
  bDebug                     bool
  bXML                       bool
  bXMLDontParseResponse      bool

  nLastStatusCode int

  RawData interface{}  // used to contain the raw response msg mody

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

  r.setUrl(url)

  r.sName = name
  r.setMethod(method)

  r.bInnerMap = false
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
// func (pRA *Restapi) DebugOn()
//
// Turn off debugging
//

func (pRA *Restapi) DebugOff(){
  pRA.bDebug = false
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
// func (pRA *Restapi) Send() bool
//
// Sends the API request
//

func (pRA *Restapi) Send() bool {

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

  req, _ := http.NewRequest(pRA.sMethodString, pRA.sUrl, nil)


  if(pRA.bRequiresAccessToken){
    req.Header.Add("Authorization", pRA.sAccessToken)
  }

//  req.Header.Add("Accept", "*/*")

  req.Header.Add("cache-control", "no-cache")
  req.Header.Add("Content-Type", "application/json")


  if(pRA.bDebug){
    fmt.Println(req)
  }

  res, err := http.DefaultClient.Do(req)

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

