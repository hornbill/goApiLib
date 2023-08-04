# Go Hornbill API lib


## Integration

Various Hornbill Integration methods are documented here: https://wiki.hornbill.com/index.php/Integration

## Documentation

```go
	package main

	import (
	        "encoding/base64"
	        "fmt"
	        "git.hornbill.com/apiLib"
	        "log"
	)

	func main() {

		    //-- apiLib.NewXmlmcInstance("testinstance") can also be used to lookup the endpoint
	        conn := apiLib.NewXmlmcInstance("https://eurapi.hornbill.com/testinstance/xmlmc/")
	        err := conn.SetParam("userId", "admin")
	        if err != nil {
	                fmt.Println(err)
	        }
	        err = conn.SetParam("password", base64.StdEncoding.EncodeToString([]byte("Password")))
	        if err != nil {
	                fmt.Println(err)
	        }

	        // SetTimeout to 40 from default 30 seconds
	        conn.SetTimeout(40)

	        //actually make the call to the server
	        result, err := conn.Invoke("session", "userLogon")
	        if err != nil {
	                log.Fatal(err)
	        }
	        fmt.Println(result)

	        // You can also check the response code to show the call really did return succesfully
	        if conn.GetStatusCode() != 200 {
	                log.Fatal("xmlmc called returned non 200 http status")
	        }

	        // GetSessionId returns the current ESPSessionId that has been set after a userLogon call.
	        sessID := conn.GetSessionID()
	        fmt.Println(sessID)

	        // Reuse the same conn
	        err = conn.SetParam("stage", "1")
	        if err != nil {
	                fmt.Println(err)
	        }

	        // We are going to set the jsonResponse flag so we get our response back as json rather than xml
	        conn.SetJSONResponse(true)
	        //invoke the command
	        result2, err := conn.Invoke("system", "pingCheck")
	        if err != nil {
	                log.Fatal(err)
	        }
	        fmt.Println(result2)
	        sessID = conn.GetSessionID()
	        fmt.Println(sessID)

	        // Set some params that we are not going to invoke
	        err = conn.SetParam("userId", "admin")
	        if err != nil {
	                fmt.Println(err)
	        }
	        err = conn.SetParam("password", base64.StdEncoding.EncodeToString([]byte("password")))
	        if err != nil {
	                fmt.Println(err)
	        }

	        //Get the currently set paramaters
	        setParams := conn.GetParam()
	        fmt.Println(setParams)

	        // Clear the paramters
	        conn.ClearParam()

	        //Get the currently set paramaters again to prove they have been cleared
	        setParams = conn.GetParam()
	        fmt.Println(setParams)

	        //Complex xml
	        // OpenElement

	        // CloseElement
	}
```