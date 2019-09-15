# Simple restapi handling tool

Created to do simple restapi calls in go.  There are probably better tool :)

# Example

Using Tesla APi to get a list of vechiles

```go
  vehicleList := restapi.NewGet("vehicles", https://owner-api.teslamotors.com/api/1/vehicles)


  vehicleList.SetBearerAccessToken("AccessTokenFromTheAuthenticationAPI")
  vehicleList.HasInnerMapArray("response","count")

  if(vehicleList.Send()){
    vehicleList.Dump()
  }else{
    fmt.Println("get vehicles list failed")
  }

  count := vehicleList.GetValueInt("count")

  fmt.Printf("Number of vehicles[%d]\n",count)

  for j:= 0; j < count; j++ {
    fmt.Println("Vehicle: ", j)
    fmt.Printf("Ids[%s]\n", vehicleList.GetArrayValueString(j,"id_s"))
    fmt.Printf("Vin[%s]\n", vehicleList.GetArrayValueString(j,"vin"))
    fmt.Printf("DisplayName[%s]\n", vehicleList.GetArrayValueString(j,"display_name"))
    fmt.Printf("State[%s]\n", vehicleList.GetArrayValueString(j,"state"))

  }
```


