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

// Notice: This code has been modified from its original source.
// Modifications are licensed as specified below.
//
// Copyright (c) 2014, fromkeith
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice, this
//   list of conditions and the following disclaimer in the documentation and/or
//   other materials provided with the distribution.
//
// * Neither the name of the fromkeith nor the names of its
//   contributors may be used to endorse or promote products derived from
//   this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
// ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
///

package gorest

import (
	"compress/gzip"
	"fmt"
	"io"
	"log" // for the default logger
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
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

var (
	defaultGet 	= 200
	defaultPost = 202
	defaultPut 	= 200
	defaultDelete = 200
	defaultHead = 200
	defaultOptions = 200
)

type endPointStruct struct {
	name                 string
	requestMethod        string
	signiture            string
	muxRoot              string
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
	overrideCharset      string // overrides what to set the charset to
	overrideProducesMime string // overrides the produces mime type
	overrideConsumesMime string // overrides the produces mime type
	allowGzip 		     int // 0 false, 1 true, 2 unitialized
}

type EndPointHelper struct {
	endPoint endPointStruct
}

type restStatus struct {
	httpCode int
	reason   string //Especially for code in range 4XX to 5XX
	callTime 	time.Duration
}

func (err restStatus) String() string {
	return err.reason
}

type serviceMetaData struct {
	template     interface{}
	consumesMime string
	producesMime string
	charset 	 string
	root         string
	realm        string
	allowGzip    bool
}

// HealthHandler reports some overal health information about requests.
type HealthHandler interface {
	// Called on the handling of a request.
	// Reports the Response (status) code of a request. Usefull to help capture
	// the overall health of your server.
	//	callDuration is the time spent inside your handler
	ReportResponseCode(urlPath *url.URL, code int, endPoint *EndPointHelper, callDuration time.Duration)
	// called when an on going call is taking too long. Call length is defined via SetWarningDuration
	ReportLongCall(endPoint *EndPointHelper, startedAt time.Time)
}

// A simple interface to wrap a basic leveled logger.
// The format strings to do not have newlines on them.
type SimpleLogger interface {
	Tracef(fmt string, args ... interface{})
	Infof(fmt string, args ... interface{})
	Warnf(fmt string, args ... interface{})
	Errorf(fmt string, args ... interface{})
	// Log the panic and exit.
	Panicf(fmt string, args ... interface{})
}

var restManager *manager
var handlerInitialised bool

// The request's response writter and request body. Along with the recover object as returned by recover()
type RecoverHandlerFunc func(http.ResponseWriter, *http.Request, interface{})

type manager struct {
	serviceTypes map[string]serviceMetaData
	endpoints    map[string]endPointStruct
	serverRecoverHandler 	RecoverHandlerFunc
	serverHealthHandler 	HealthHandler
	logger 					SimpleLogger
	callDurationWarning 	time.Duration
}

type defaultLogger struct {}

func (d defaultLogger) Tracef(fmt string, args ... interface{}) {

}

func (d defaultLogger) Infof(fmt string, args ... interface{}) {
	log.Printf(fmt + "\n", args...)
}
func (d defaultLogger) Warnf(fmt string, args ... interface{}) {
	log.Printf(fmt + "\n", args...)
}
func (d defaultLogger) Errorf(fmt string, args ... interface{}) {
	log.Printf(fmt + "\n", args...)
}
func (d defaultLogger) Panicf(fmt string, args ... interface{}) {
	log.Panicf(fmt + "\n", args...)
}


func newManager() *manager {
	man := new(manager)
	man.serviceTypes = make(map[string]serviceMetaData, 0)
	man.endpoints = make(map[string]endPointStruct, 0)
	man.logger = defaultLogger{}
	man.callDurationWarning = -1
	return man
}
func init() {
	RegisterMarshaller(Application_Json, NewJSONMarshaller())
}

// Name of the endpoint. The variable name you used to in the endpoint definition
func (e EndPointHelper) GetName() string {
	return e.endPoint.name
}
// The signature of the path used in the endpoint defintion. Eg: /cars/{id:string}
func (e EndPointHelper) GetSignature() string {
	return e.endPoint.signiture
}
// The method used on the request. Eg. GET, PUT, ...
func (e EndPointHelper) GetMethod() string {
	return e.endPoint.requestMethod
}

//Registeres a service on the rootpath.
//See example below:
//
//	package main
//	import (
// 	   "code.google.com/p/gorest"
//	        "http"
//	)
//	func main() {
//	    gorest.RegisterService(new(HelloService)) //Register our service
//	    http.Handle("/",gorest.Handle())
//	    http.ListenAndServe(":8787",nil)
//	}
//
//	//Service Definition
//	type HelloService struct {
//	    gorest.RestService `root:"/tutorial/"`
//	    helloWorld  gorest.EndPoint `method:"GET" path:"/hello-world/" output:"string"`
//	    sayHello    gorest.EndPoint `method:"GET" path:"/hello/{name:string}" output:"string"`
//	}
//	func(serv HelloService) HelloWorld() string{
// 	   return "Hello World"
//	}
//	func(serv HelloService) SayHello(name string) string{
//	    return "Hello " + name
//	}
func RegisterService(h interface{}) {
	RegisterServiceOnPath("", h)
}

//Registeres a service under the specified path.
//See example below:
//
//	package main
//	import (
//	    "code.google.com/p/gorest"
//	        "http"
//	)
//	func main() {
//	    gorest.RegisterServiceOnPath("/rest/",new(HelloService)) //Register our service
//	    http.Handle("/",gorest.Handle())
//	    http.ListenAndServe(":8787",nil)
//	}
//
//	//Service Definition
//	type HelloService struct {
//	    gorest.RestService `root:"/tutorial/"`
//	    helloWorld  gorest.EndPoint `method:"GET" path:"/hello-world/" output:"string"`
//	    sayHello    gorest.EndPoint `method:"GET" path:"/hello/{name:string}" output:"string"`
//	}
//	func(serv HelloService) HelloWorld() string{
//	    return "Hello World"
//	}
//	func(serv HelloService) SayHello(name string) string{
//	    return "Hello " + name
//	}
func RegisterServiceOnPath(root string, h interface{}) {
	//We only initialise the handler management once we know gorest is being used to hanlde request as well, not just client.
	intializeManager()

	if root == "/" {
		root = ""
	}

	if root != "" {
		root = strings.Trim(root, "/")
		root = "/" + root
	}

	registerService(root, h)
}

//ServeHTTP dispatches the request to the handler whose pattern most closely matches the request URL.
func (_ manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	url_, err := url.QueryUnescape(r.URL.RequestURI())

	_manager().logger.Tracef("ServeHttp [%s] %s", r.Method, r.URL.String())

	var knownEndpoint *endPointStruct

	defer func() {
		if rec := recover(); rec != nil {
			recoverFunc := _manager().serverRecoverHandler
			if recoverFunc != nil {
				recoverFunc(w, r, rec)
				if _manager().serverHealthHandler != nil {
					var epHelp *EndPointHelper
					if knownEndpoint != nil {
						epHelp = &EndPointHelper{*knownEndpoint}
					}
					_manager().serverHealthHandler.ReportResponseCode(r.URL, 500, epHelp, 0)
				}
			} else if _manager().logger != nil {
				_manager().logger.Errorf("Internal Server Error: Could not serve page: %s %s", r.Method, url_)
				_manager().logger.Errorf("Panic: %v", rec)
				_manager().logger.Errorf("%s", debug.Stack())
			}
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	if err != nil {
		_manager().logger.Errorf("Could not serve page: %s %v Error: %v", r.Method, r.URL.RequestURI(), err)
		w.WriteHeader(400)
		w.Write([]byte("Client sent bad request."))
		if _manager().serverHealthHandler != nil {
			_manager().serverHealthHandler.ReportResponseCode(r.URL, 400, nil, 0)
		}
		return
	}

	if ep, args, queryArgs, xsrft, found := getEndPointByUrl(r.Method, url_); found {

		knownEndpoint = &ep

		if xsrft == "" {
			if c, err := r.Cookie(XSXRF_COOKIE_NAME); err == nil {
				xsrft = c.Value
			}
		}

		ctx := new(Context)
		ctx.writer = w
		ctx.request = r
		ctx.args = args
		ctx.queryArgs = queryArgs
		ctx.xsrftoken = xsrft

		data, state := prepareServe(ctx, ep)
		writtenStatusCode := -1

		var mimeType string
		if mimeType = ep.overrideProducesMime; mimeType == "" {
			mimeType = _manager().getType(ep.parentTypeName).producesMime
		}
		var charset string
		if charset = ep.overrideCharset; charset == "" {
			charset = _manager().getType(ep.parentTypeName).charset
		}
		if charset != "-" && charset != "" {
			mimeType = fmt.Sprintf("%s; charset=%s", mimeType, charset)
		}


		if state.httpCode == http.StatusOK {
			switch ep.requestMethod {
			case POST, PUT, DELETE, HEAD, OPTIONS:
				{
					if ctx.responseCode == 0 {
						writtenStatusCode = getDefaultResponseCode(ep.requestMethod)
					} else {
						if !ctx.dataHasBeenWritten {
							writtenStatusCode = ctx.responseCode
						}
					}
				}
			case GET:
				{
					if ctx.responseCode == 0 {
						writtenStatusCode = getDefaultResponseCode(ep.requestMethod)
					} else {
						if !ctx.dataHasBeenWritten {
							writtenStatusCode = ctx.responseCode
						}
					}
				}
			}

			if !ctx.responseMimeSet &&  data != nil {
				w.Header().Set("Content-Type", mimeType)
			}

			if data != nil && !ctx.overide {
				if ep.allowGzip == 1 && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
					w.Header().Set("Content-Encoding", "gzip")
					w.WriteHeader(writtenStatusCode)
					gzipWriter := gzip.NewWriter(w)
					defer gzipWriter.Close()
					io.Copy(gzipWriter, data)
				} else {
					w.WriteHeader(writtenStatusCode)
					io.Copy(w, data)
				}
			} else {
				w.WriteHeader(writtenStatusCode)
			}

		} else {
			if _manager().logger != nil {
				if state.httpCode >= 500 {
					_manager().logger.Errorf("Problem with request. Error: %s %d %s; Request: %v", r.Method, state.httpCode, state.reason, r.URL.RequestURI())
				} else {
					_manager().logger.Warnf("Problem with request. Error: %s %d %s; Request: %v", r.Method, state.httpCode, state.reason, r.URL.RequestURI())
				}
			}
			writtenStatusCode = state.httpCode
			w.WriteHeader(state.httpCode)
			w.Write([]byte(state.reason))
		}

		if _manager().serverHealthHandler != nil && writtenStatusCode != -1 {
			go _manager().serverHealthHandler.ReportResponseCode(r.URL, writtenStatusCode, &EndPointHelper{ep}, state.callTime)
		}

	} else {
		if _manager().logger != nil {
			_manager().logger.Warnf("Could not serve page, path not found: %s %s", r.Method, url_)
		}
		//		println("Could not serve page, path not found: ", r.Method, url_)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("The resource in the requested path could not be found."))
	}
}

func intializeManager() {
	if !handlerInitialised {
		restManager = newManager()
		handlerInitialised = true
	}
}

// Overrides the logger with your own version
func OverrideLogger(logger SimpleLogger) {
	intializeManager()
	_manager().logger = logger
}

// Registers the handler to deal with health information
func RegisterHealthHandler(handler HealthHandler) {
	intializeManager()
	_manager().serverHealthHandler = handler
}

// Register a callback that will be fired whenever gorest runs into a runtime error.
func RegisterRecoveryHandler(handler RecoverHandlerFunc) {
	intializeManager()
	_manager().serverRecoverHandler = handler
}

// If a handler is still handling a request after the specified duration,
// the health handler's ReportLongCall method will be invoked
func SetWarningDuration(dur time.Duration) {
	intializeManager()
	_manager().callDurationWarning = dur
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

//Registeres the function to be used for handling all requests directed to gorest.
func HandleFunc(w http.ResponseWriter, r *http.Request) {
	if _manager().logger != nil {
		_manager().logger.Infof("Serving URL : %s %v", r.Method, r.URL.RequestURI())
	}
	defer func() {
		if rec := recover(); rec != nil {
			recoverFunc := _manager().serverRecoverHandler
			if recoverFunc != nil {
				recoverFunc(w, r, rec)
			} else if _manager().logger != nil {
				_manager().logger.Errorf("Internal Server Error: Could not serve page: %s %s", r.Method, r.URL.Path)
				_manager().logger.Errorf("Panic: %v", rec)
				_manager().logger.Errorf("%s", debug.Stack())
			}
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	restManager.ServeHTTP(w, r)
}

//Runs the default "net/http" DefaultServeMux on the specified port.
//All requests are handled using gorest.HandleFunc()
func ServeStandAlone(port int) {
	http.HandleFunc("/", HandleFunc)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

func _manager() *manager {
	return restManager
}
func Handle() manager {
	return *restManager
}

func SetDefaultResponseCode(method string, val int) {
	switch method {
	case GET:
		defaultGet = val
	case POST:
		defaultPost = val
	case PUT:
		defaultPut = val
	case DELETE:
		defaultDelete = val
	case HEAD:
		defaultHead = val
	case OPTIONS:
		defaultOptions = val
	}
}

func getDefaultResponseCode(method string) int {
	switch method {
	case GET:
		return defaultGet
	case POST:
		return defaultPost
	case PUT:
		return defaultPut
	case DELETE:
		return defaultDelete
	case HEAD:
		return defaultHead
	case OPTIONS:
		return defaultOptions
	}

	return 200
}
