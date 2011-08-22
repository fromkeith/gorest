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

import (
	"http"
	"strconv"
)

type GoRestService interface {
	ResponseBuilder() *ResponseBuilder
}

const (
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
)

type endPointStruct struct {
	name                 string
	requestMethod        string
	signiture            string
	root                 string
	nonParamPathPart     map[int]string
	params               []param //path parameter name and position
	queryParams          []param
	signitureLen         int
	paramLen             int
	inputMime            string
	outputType           string
	outputTypeIsArray    bool
	outputTypeIsMap      bool
	postdataType         string
	postdataTypeIsArray  bool
	postdataTypeIsMap    bool
	isVariableLength     bool
	parentTypeName       string
	methodNumberInParent int
	role                 string
}

type restStatus struct {
	httpCode int
	reason   string //Especially for code in range 4XX to 5XX
}

func (err restStatus) String() string {
	return err.reason
}


type serviceMetaData struct {
	template     interface{}
	consumesMime string
	producesMime string
	root         string
	realm        string
}


var restManager *manager
var handlerInitialised bool

type manager struct {
	serviceTypes map[string]serviceMetaData
	endpoints    map[string]endPointStruct
}

func newManager() *manager {
	man := new(manager)
	man.serviceTypes = make(map[string]serviceMetaData, 0)
	man.endpoints = make(map[string]endPointStruct, 0)

	return man
}
func init() {
	RegisterMarshaller(Application_Json, NewJSONMarshaller())

}


func RegisterService(h interface{}) {
	//We only initialise the handler management once we know gorest is being used to hanlde request as well, not just client.
	if !handlerInitialised {
		restManager = newManager()
		handlerInitialised = true
	}

	registerService(h)
}


func (man *manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url,err := http.URLUnescape(r.RawURL)
	if err!=nil{
		println("Could not serve page: ", r.RawURL)
		w.WriteHeader(400)
		w.Write([]byte("Client sent bad request."))
		return
	}
	if ep, args, queryArgs, found := getEndPointByUrl(r.Method, url); found {

		ctx := new(Context)
		ctx.writer = w
		ctx.request = r
		ctx.args = args
		ctx.queryArgs = queryArgs

		data, state := prepareServe(ctx, ep)

		if state.httpCode == http.StatusOK {
			switch ep.requestMethod {
			case POST, PUT, DELETE,HEAD,OPTIONS:
				{
					if ctx.responseCode == 0 {
						w.WriteHeader(getDefaultResponseCode(ep.requestMethod))
					} else {
						if !ctx.dataHasBeenWritten {
							w.WriteHeader(ctx.responseCode)
						}
					}
				}
			case GET:
				{
					if ctx.responseCode == 0 {
						if !ctx.responseMimeSet {
							w.Header().Set("Content-Type", man.getType(ep.parentTypeName).producesMime)
						}
						w.WriteHeader(getDefaultResponseCode(ep.requestMethod))
					} else {
						if !ctx.dataHasBeenWritten {
							if !ctx.responseMimeSet {
								w.Header().Set("Content-Type", man.getType(ep.parentTypeName).producesMime)
							}
							w.WriteHeader(ctx.responseCode)
						}
					}

					if !ctx.overide {
						w.Write(data)
					}

				}
			}

		} else {
			w.WriteHeader(state.httpCode)
			w.Write([]byte(state.reason))
		}

	} else {
		println("Could not serve page: ", url)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("The resource in the requested path could not be found."))
	}

}

func (man *manager) getType(name string) serviceMetaData {

	return man.serviceTypes[name]
}
func (man *manager) addType(name string, i serviceMetaData) string {
	for str, _ := range man.serviceTypes {
		if name == str {
			return str
		}
	}

	man.serviceTypes[name] = i
	return name
}
func (man *manager) addEndPoint(ep endPointStruct) {
	man.endpoints[ep.requestMethod+":"+ep.signiture] = ep
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	//    log.Println("Serving URL : ",r.RawURL)
	restManager.ServeHTTP(w, r)
}

func ServeStandAlone(port int) {
	http.HandleFunc("/", mainHandler)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

func _manager() *manager {
	return restManager
}


func getDefaultResponseCode(method string) int {
	switch method {
	case GET, PUT, DELETE:
		{
			return 200
		}
	case POST:
		{
			return 202
		}
	default:
		{
			return 200
		}
	}

	return 200
}
