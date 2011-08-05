//Copyright 2011 Siyabonga Dlamini (siyabonga.dlamini@gmail.com). All rights reserved.
//
//Redistribution and use in source and binary forms, with or without
//modification, are permitted provided that the following conditions
//are met:
//
//  1. Redistributions of source code must retain the above copyright
//     notice, this list of conditions and the following disclaimer.
//
//  2. Redistributions in binary form must reproduce the above copyright
//     notice, this list of conditions and the following disclaimer
//     in the documentation and/or other materials provided with the
//     distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE AUTHOR ``AS IS'' AND ANY EXPRESS OR
//IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
//OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
//IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
//SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
//PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS;
//OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY,
//WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR
//OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF
//ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package gorest
import "log"
import "http"
import "strconv"
import "json"
import "os"

type GoRestService interface{
    ResponseBuilder()*ResponseBuilder
}

const(
    GET="GET"
    POST="POST"
    PUT="PUT"
    DELETE="DELETE"
    OPTIONS="OPTIONS"
)

type endPointStruct struct{
    name string
    requestMethod string
    signiture string
	root string
    params map[int]param   //path parameter name and position

    signitureLen int
    paramLen int

    inputMime  string
    outputType string
    outputTypeIsArray bool
    outputTypeIsMap bool
    postdataType string

    parentTypeName string
    methodNumberInParent int
}

type restStatus struct{
    httpCode int
    reason string //Especially for code in range 4XX to 5XX
}

func(err restStatus) String()string{
    return err.reason
}




var restManager *manager

func init(){
      restManager=new(manager)
      restManager.serviceTypes = make(map[string]serviceMetaData,0)
      restManager.endpoints = make(map[string]endPointStruct,0)
}



type manager struct{
    serviceTypes map[string]serviceMetaData
    endpoints map[string]endPointStruct
}

type serviceMetaData struct{
    template interface{}
    consumesMime string
    producesMime string
    root string
}

func(man *manager) ServeHTTP(w http.ResponseWriter,r *http.Request){
    if ep,args,found:=getEndPointByUrl(r.Method,r.RawURL);found{

        ctx:=new(Context)
        ctx.writer=w
        ctx.request=r
        ctx.args =args

        data,state:= prepareServe(ctx,ep)

        if state.httpCode == http.StatusOK{
            switch ep.requestMethod{
                case POST,PUT,DELETE:{
                    if ctx.responseCode ==0{
                        w.WriteHeader(getDefaultResponseCode(ep.requestMethod))
                    }else{
                        if !ctx.dataHasBeenWritten{
                            w.WriteHeader(ctx.responseCode)
                        }
                    }
                }
                case GET:{
                    if ctx.responseCode ==0{
                        w.WriteHeader(getDefaultResponseCode(ep.requestMethod))
                    }else{
                        if !ctx.dataHasBeenWritten{
                            w.WriteHeader(ctx.responseCode)
                        }
                    }

                    if !ctx.overide{
                        w.Write(data)
                    }

//                    if data !=nil{
//                      w.Header().Set("Content-Type", "application/json") //TODO: Use registered mime-type here
//                      w.Header().Set("Connection", "Keep-Alive")
////                    w.Header().Set("Content-length",string(len(data)))
//                      w.WriteHeader(http.StatusOK)
//                      w.Write(data)
//                    }

                }
            }

        }else{
            w.WriteHeader(state.httpCode)
            w.Write([]byte(state.reason))
        }

    }else{
        println("Could not serve page: ", r.RawURL)
        w.WriteHeader(http.StatusNotFound)
        w.Write([]byte("The resource in the requested path could not be found."))
    }

}

func(man *manager) getType(name string) serviceMetaData{
     for str,_:=range man.serviceTypes{
        println("service name:",str)
     }
     return man.serviceTypes[name]
}
func(man *manager) addType(name string,i serviceMetaData) string{
   for str,_:=range man.serviceTypes{
        if name == str{
            return str
        }
   }

   man.serviceTypes[name]=i
   return name
}
func(man *manager) addEndPoint(ep endPointStruct){
    man.endpoints[ep.requestMethod+":"+ep.signiture] = ep
}

func mainHandler(w http.ResponseWriter, r *http.Request){
    log.Println("Serving URL : ",r.RawURL)
    restManager.ServeHTTP(w,r)
}

func ServeStandAlone(port int){
    http.HandleFunc("/", mainHandler)
    http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

func _manager()*manager{
    return restManager
}

func unMarshal(mime string,data []byte, i interface{}) os.Error{
    //TODO: Get the appropriate marshaller from list of registered ones
    if mime == "application/json"{
      return json.Unmarshal(data,i)
    }
    println("Could not find any registered marshaller for mime "+mime)
    return os.NewError("Could not find any registered marshaller for mime "+mime)
}

func marshal(mime string,i interface{})([]byte,os.Error){
     if mime == "application/json"{
        return json.Marshal(i)
     }
     println("Could not find any registered marshaller for mime "+mime)
     return nil,os.NewError("Could not find any registered marshaller for mime "+mime)
}
func getDefaultResponseCode(method string)int{
    switch method{
        case GET, PUT,DELETE:{
            return 200
        }
        case POST:{
            return 202
        }
        default:{
            return 200
        }
    }

    return 200
}






