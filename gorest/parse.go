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
	"strings"
	"log"
	"reflect"
)


type argumentData struct {
	parameter param
	data      string
}
type param struct {
	positionInPath int
	name           string
	typeName       string
}

var ALLOWED_PAR_TYPES = []string{"string", "int", "bool", "float32", "float64"}


func prepServiceMetaData(tags reflect.StructTag, i interface{}) serviceMetaData {
	md := new(serviceMetaData)

	if tag := tags.Get("root"); tag != "" {
		md.root = tag
	}

	if tag := tags.Get("consumes"); tag != "" {
		md.consumesMime = tag
	} else {
		md.consumesMime = Application_Json //Default	
	}
	if tag := tags.Get("produces"); tag != "" {
		md.producesMime = tag
	} else {
		md.consumesMime = Application_Json //Default	
	}

	md.template = i
	return *md
}

func makeEndPointStruct(tags reflect.StructTag, serviceRoot string) endPointStruct {

	ms := new(endPointStruct)

	if tag := tags.Get("method"); tag != "" {
		if tag == "GET" {
			ms.requestMethod = GET
		} else if tag == "POST" {
			ms.requestMethod = POST
		} else if tag == "PUT" {
			ms.requestMethod = PUT
		} else if tag == "DELETE" {
			ms.requestMethod = DELETE
		} else if tag == "OPTIONS" {
			ms.requestMethod = OPTIONS
		} else {
			log.Panic("Unknown method type:[" + tag + "] in endpoint declaration. Allowed types {GET,POST,PUT,DELETE,OPTIONS}")
		}

		if tag := tags.Get("path"); tag != "" {
			serviceRoot = strings.TrimRight(serviceRoot, "/")
			ms.signiture = serviceRoot + "/" + strings.Trim(tag, "/")
		} else {
			panic("Endpoint declaration must have the tags 'method' and 'path' ")
		}

		if tag := tags.Get("output"); tag != "" {
			ms.outputType = tag
			if strings.HasPrefix(tag, "[]") { //Check for slice/array/list types.
				ms.outputTypeIsArray = true
				ms.outputType = ms.outputType[2:]
			}
			if strings.HasPrefix(tag, "map[") { //Check for map[string]. We only handle string keyed maps!!!

				if ms.outputType[4:10] == "string" {
					ms.outputTypeIsMap = true
					ms.outputType = ms.outputType[11:]
				} else {
					panic("Only string keyed maps e.g( map[string]... ) are allowed on the [output] tag. Endpoint: " + ms.signiture)
				}

			}
		}

		if tag := tags.Get("input"); tag != "" {
			ms.inputMime = tag
		}

		if tag := tags.Get("postdata"); tag != "" {
			ms.postdataType = tag
			if strings.HasPrefix(tag, "[]") { //Check for slice/array/list types.
				ms.postdataTypeIsArray = true
				ms.postdataType = ms.postdataType[2:]
			}
			if strings.HasPrefix(tag, "map[") { //Check for map[string]. We only handle string keyed maps!!!

				if ms.postdataType[4:10] == "string" {
					ms.postdataTypeIsMap = true
					ms.postdataType = ms.postdataType[11:]
				} else {
					panic("Only string keyed maps e.g( map[string]... ) are allowed on the [postdata] tag. Endpoint: " + ms.signiture)
				}

			}
		}

		parseParams(ms)
		return *ms
	}

	panic("Endpoint declaration must have the tags 'method' and 'path' ")

}
func parseParams(e *endPointStruct) {
	e.signiture = strings.Trim(e.signiture, "/")
	e.params = make(map[int]param, 0)

	i := strings.Index(e.signiture, "{")
	if i > 0 {
		e.root = e.signiture[0:i]
	} else {
		e.root = e.signiture
	}

	//Extract path parameters
	for pos, str1 := range strings.Split(e.signiture, "/") {
		e.signitureLen++
		if strings.HasPrefix(str1, "{") && strings.HasSuffix(str1, "}") {

			temp := strings.Trim(str1, "{}")
			ind := 0
			if ind = strings.Index(temp, ":"); ind == -1 {
				panic("Please ensure that parameter names(" + temp + ") have associated types in REST path: " + e.signiture)
			}
			parName := temp[:ind]
			typeName := temp[ind+1:]

			if !isAllowedParamType(typeName) {
				panic("Type " + typeName + " is not allowed for path-parameters in REST path: " + e.signiture)
			}

			for _, par := range e.params {
				if par.name == parName {
					panic("Duplicate field name(" + parName + ") in REST path: " + e.signiture)
				}
			}

			e.params[e.paramLen] = param{pos, parName, typeName}
			e.paramLen++
		}
	}

	if ep, there := _manager().endpoints[e.signiture]; there && ep.requestMethod == e.requestMethod {
		panic("Endpoint already registered: " + e.signiture)
	}

	for _, ep := range _manager().endpoints {
		if ep.root == e.root && ep.signitureLen == e.signitureLen && ep.requestMethod == e.requestMethod {

			panic("Can not register two endpoints with same request-method(" + ep.requestMethod + "), same root and same amount of parameters: " + e.signiture)
		}
	}
}

func isAllowedParamType(typeName string) bool {
	for _, s := range ALLOWED_PAR_TYPES {
		if s == strings.ToLower(typeName) {
			return true
		}
	}
	return false
}


func getEndPointByUrl(method string, url string) (endPointStruct, []argumentData, bool) {

	url = strings.Trim(url, "/")
	totalParts := strings.Count(url, "/")
	totalParts++

	epRet := new(endPointStruct)
	adarr := make([]argumentData, 0)
	ok := false
	for _, ep := range _manager().endpoints {

		if strings.Contains(url, ep.root) && ep.signitureLen == totalParts && ep.requestMethod == method { //TODO: Make sure it starts with
			//           log.Println("End point found: ",ep.requestMethod,ep.root,ep.signiture,ep.signitureLen,url,ep.paramLen)
			for _, par := range ep.params {
				for upos, str1 := range strings.Split(url, "/") {
					if par.positionInPath == upos {
						adarr = append(adarr, argumentData{par, strings.Trim(str1, " ")})
					}
				}
			}
			ok = true
			return ep, adarr, ok
		}
	}
	return *epRet, adarr, ok
}
