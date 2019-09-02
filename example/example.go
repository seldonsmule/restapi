package main

import (
	"os"
	"fmt"
        "restapi"
        //"time"
//        "bufio"
        //"syscall"
 //       "strconv"
  //      "strings"
//	"net/http"
//	"io/ioutil"
 //       "encoding/json"
//        "database/sql"
  //      "time"
//        _ "github.com/mattn/go-sqlite3" 
        "github.com/seldonsmule/logmsg"
//        "golang.org/x/crypto/ssh/terminal"
)

func bigarray(bDebug bool){

  r := restapi.NewGet("Chargering","http://localhost:3000/nearby_charging_sites")

  r.SetBearerAccessToken("accessTokeng2342xxx")

  if(bDebug){
    r.DebugOn()
  }


  r.HasInnerMap("response")

  if(r.Send()){

    r.Dump()

  }

  tmp1 := r.GetValue("destination_charging")

  fmt.Printf("destination_charging[%s]\n", tmp1)

  myarray := r.CastArray(tmp1)

  fmt.Println("array len:", len(myarray))

  for i:=0; i < len(myarray); i++ {

    tmpmap := r.CastMap(myarray[i])

    for name, value := range tmpmap{
  
      fmt.Println(name, "=", value)

    }

  }

/*
  mymap := dest.(map[string]interface{})
 
  for name, value := range mymap{
  
    fmt.Println(name, "=", value)

  }
*/

}

func simple(bDebug bool){

  r := restapi.NewGet("authentication","http://localhost:3000/authentication")

  if(bDebug){
    r.DebugOn()
  }


  if(r.Send()){

    r.Dump()

  }

  fmt.Printf("access token[%s]\n", r.GetValue("access_token"))
}

func sunriseset(bDebug bool){


  r := restapi.NewGet("sunriseset", "https://weather.cit.api.here.com/weather/1.0/report.json?product=forecast_astronomy&name=DC&app_id=DemoAppId01082013GAL&app_code=AJKnXv84fjrb0KIHawS0Tg")


  if(bDebug){
    r.DebugOn()
  }

  r.HasInnerMap("astronomy")

  if(r.Send()){

    r.Dump()

  }

  fmt.Printf("--------------------------\n")

  fmt.Printf("astronomy[%s]\n", r.GetValue("astronomy"))

  fmt.Printf("--------------------------\n")

// here is the deal - this stupid thing is an array of maps, not 
// something considered in my original Telsa use case.  here is how to
// get the data...

// 1. Get the info as an array

  astroArray := r.CastArray(r.GetValue("astronomy"))

  fmt.Printf("ArrayLength[%d]\n", len(astroArray))

  for k, v := range astroArray {
    fmt.Println(k, "=", v)

  } // end for loop

// 2. Get the map desired

  fmt.Printf("--------------------------\n")
  astroMap := r.CastMap(astroArray[1])
  for k, v := range astroMap {
    fmt.Println(k, "=", v)
  } // end for loop

// 3. get teh value in the map
  fmt.Printf("--------------------------\n")
  fmt.Printf("sunset[%s]\n", astroMap["sunset"])
  fmt.Printf("sunrise[%s]\n", astroMap["sunrise"])


}

func innermap(bDebug bool){

  r := restapi.NewGet("authentication","http://localhost:3000/charge_state")

  r.SetBearerAccessToken("accessTokeng2342xxx")

  if(bDebug){
    r.DebugOn()
  }

  r.HasInnerMap("response")


  if(r.Send()){

    r.Dump()

  }

  fmt.Printf("Batter Level[%f]\n", r.GetValue("battery_level"))
}


func maparray(bDebug bool){

  //var m MyResp

  r := restapi.NewGet("authentication","http://localhost:3000/vehicles")

  r.SetBearerAccessToken("accessTokeng2342xxx")

  if(bDebug){
    r.DebugOn()
  }

  r.HasInnerMapArray("response", "count")


  if(r.Send()){

    r.Dump()

  }

  fmt.Println("Calling GetArrayValue Index 0 vin:", r.GetArrayValue(0, "vin"))
}

func help(){

  fmt.Println("usage example test_name [debug]")
  fmt.Println()
  fmt.Println("simple - straight forward json response")
  fmt.Println("innermap - handles a json response that is layered")
  fmt.Println("maparray - handles a json response that is layered and is an array of maps")
  fmt.Println("bigparray - lots of data")
  fmt.Println("sunriseset - sunrise/sunset example data")


}

func main() {

  bDebug := false

  fmt.Println("starting example for restapi")
  logmsg.SetLogFile("example.log");

  args := os.Args

  if(len(args) < 2){
    help()
    os.Exit(1)
  }

  if(len(args) == 3){
  
    switch args[2] {
      case "debug":
        bDebug = true
        fmt.Println("Debug on")

      default:
        help()
        os.Exit(2)
    }
  }

  switch args[1] {

    case "simple":
      simple(bDebug)

    case "sunriseset":
      sunriseset(bDebug)

    case "innermap":
      innermap(bDebug)

    case "maparray":
      maparray(bDebug)

    case "bigarray":
      bigarray(bDebug)

    default:
      help()

  }

}
