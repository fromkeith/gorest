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

type EndPoint bool


type RestService struct {
	Context *Context
}


func (serv RestService) ResponseBuilder() *ResponseBuilder {
	r := ResponseBuilder{ctx: serv.Context}
	return &r
}

type Context struct {
	writer  http.ResponseWriter
	request *http.Request
	args    map[string]string
	queryArgs    map[string]string

	//Response flags
	overide            bool
	responseCode       int
	responseMimeSet    bool
	dataHasBeenWritten bool
	
}

func (c *Context) Request() *http.Request {
	return c.request
}


type ResponseBuilder struct {
	ctx *Context
}

func (this *ResponseBuilder) writer() http.ResponseWriter {
	return this.ctx.writer
}
func (this *ResponseBuilder) SetResponseCode(code int) *ResponseBuilder {
	this.ctx.responseCode = code
	return this
}

func (this *ResponseBuilder) SetContentType(mime string) *ResponseBuilder {
	this.ctx.responseMimeSet = true
	this.writer().Header().Set("Content-Type", mime)
	return this
}


//Entity related
func (this *ResponseBuilder) Overide(overide bool) {
	this.ctx.overide = overide
}
func (this *ResponseBuilder) WriteAndOveride(data []byte) *ResponseBuilder {
	this.ctx.overide = true
	return this.Write(data)
}
func (this *ResponseBuilder) WriteAndContinue(data []byte) *ResponseBuilder {
	this.ctx.overide = false
	return this.Write(data)
}

func (this *ResponseBuilder) Write(data []byte) *ResponseBuilder {
	if this.ctx.responseCode == 0 {
		this.SetResponseCode(getDefaultResponseCode(this.ctx.request.Method))

	}
	if !this.ctx.dataHasBeenWritten {
		//TODO: Check for content type set.......
		this.writer().WriteHeader(this.ctx.responseCode)
	}

	this.writer().Write(data)
	this.ctx.dataHasBeenWritten = true
	return this
}


func (this *ResponseBuilder) LongPoll(delay int, producer func(interface{}) interface{}) *ResponseBuilder {

	return this
}


//Cache related
func (this *ResponseBuilder) CachePublic() *ResponseBuilder {
	this.setCache("public")
	return this
}
func (this *ResponseBuilder) CachePrivate() *ResponseBuilder {
	this.setCache("private")
	return this
}
func (this *ResponseBuilder) CacheNoCache() *ResponseBuilder {
	this.setCache("no-cache")
	return this
}
func (this *ResponseBuilder) CacheNoStore() *ResponseBuilder {
	this.setCache("no-store")
	return this
}
func (this *ResponseBuilder) CacheNoTransform() *ResponseBuilder {
	this.setCache("no-transform")
	return this
}
func (this *ResponseBuilder) CacheMustReval() *ResponseBuilder {
	this.setCache("must-revalidate")
	return this
}
func (this *ResponseBuilder) CacheProxyReval() *ResponseBuilder {
	this.setCache("proxy-revalidate")
	return this
}
func (this *ResponseBuilder) CacheMaxAge(seconds int) *ResponseBuilder {
	this.setCache("max-age = " + strconv.Itoa(seconds))
	return this
}
func (this *ResponseBuilder) CacheSMaxAge(seconds int) *ResponseBuilder {
	this.setCache("s-maxage = " + strconv.Itoa(seconds))
	return this
}
func (this *ResponseBuilder) CacheClearAllOptions() *ResponseBuilder {
	this.writer().Header().Del("Cache-control")
	return this
}


func (this *ResponseBuilder) ConnectionKeepAlive() *ResponseBuilder {
	this.writer().Header().Set("Connection", "keep-alive")
	return this
}
func (this *ResponseBuilder) ConnectionClose() *ResponseBuilder {
	this.writer().Header().Set("Connection", "close")
	return this
}
func (this *ResponseBuilder) Location(location string) *ResponseBuilder {
	this.writer().Header().Set("Location", location)
	return this
}
func (this *ResponseBuilder) Created(location string) *ResponseBuilder {
	this.ctx.responseCode = 201
	this.writer().Header().Set("Location", location)
	return this
}
func (this *ResponseBuilder) MovedPermanently(location string) *ResponseBuilder {
	this.ctx.responseCode = 301
	this.writer().Header().Set("Location", location)
	return this
}
func (this *ResponseBuilder) Found(location string) *ResponseBuilder {
	this.ctx.responseCode = 302
	this.writer().Header().Set("Location", location)
	return this
}
func (this *ResponseBuilder) SeeOther(location string) *ResponseBuilder {
	this.ctx.responseCode = 303
	this.writer().Header().Set("Location", location)
	return this
}
func (this *ResponseBuilder) MovedTemporarily(location string) *ResponseBuilder {
	this.ctx.responseCode = 307
	this.writer().Header().Set("Location", location)
	return this
}


func (this *ResponseBuilder) Age(seconds int) *ResponseBuilder {
	this.writer().Header().Set("Age", strconv.Itoa(seconds))
	return this
}
func (this *ResponseBuilder) ETag(tag string) *ResponseBuilder {
	this.writer().Header().Set("ETag", tag)
	return this
}
func (this *ResponseBuilder) Allow(tag string) *ResponseBuilder {
	this.writer().Header().Add("Allow", tag)
	return this
}

func (this *ResponseBuilder) setCache(option string) {
	this.writer().Header().Add("Cache-control", option)
}
