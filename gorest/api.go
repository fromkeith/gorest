/**
 * Created by IntelliJ IDEA.
 * User: DlaminiSi
 * Date: 2011/07/27
 * Time: 1:34 AM
 * To change this template use File | Settings | File Templates.
 */


package gorest

import(
    "http"

)

type EndPoint bool



type RestService struct{
    Context *Context
}


func (serv RestService) ResponseBuilder()*ResponseBuilder{
    r:=ResponseBuilder{ctx:serv.Context}
     return &r
}

type Context struct{
    writer http.ResponseWriter
    request *http.Request
    args []argumentData

    //Response flags
    overide bool
    responseCode int
    responseMimeSet bool
    dataHasBeenWritten bool
}

func (c *Context) Request() (*http.Request){
    return c.request
}


type Response struct{

}

type ResponseBuilder struct{
    ctx *Context
}
func(this *ResponseBuilder) writer() http.ResponseWriter{
    return this.ctx.writer
}
func(this *ResponseBuilder) SetResponseCode(code int) *ResponseBuilder{
    this.ctx.responseCode =code
    return this
}

func(this *ResponseBuilder) SetContentType(mime string) *ResponseBuilder{
    this.writer().Header().Set("Content-Type", mime)
    return this
}
func(this *ResponseBuilder) ContentType(mime string) *ResponseBuilder{
    this.writer().Header().Add("Content-Type", mime)
    return this
}

//Entity related
func(this *ResponseBuilder) Overide(overide bool){
    this.ctx.overide =overide
}
func(this *ResponseBuilder) WriteAndOveride(data []byte) *ResponseBuilder{
    this.ctx.overide =true
    return this.Write(data)
}
func(this *ResponseBuilder) WriteAndContinue(data []byte) *ResponseBuilder{
    this.ctx.overide =false
    return this.Write(data)
}
func(this *ResponseBuilder) Write(data []byte) *ResponseBuilder{
    if this.ctx.responseCode ==0{
        this.SetResponseCode(getDefaultResponseCode(this.ctx.request.Method))

    }
    if !this.ctx.dataHasBeenWritten{
       this.writer().WriteHeader(this.ctx.responseCode)
    }

    this.writer().Write(data)
    this.ctx.dataHasBeenWritten =true
    return this
}


func(this *ResponseBuilder) LongPoll(delay int,producer func(interface{})interface{}) *ResponseBuilder{

    return this
}

func(this *ResponseBuilder) Send() Response{
    return Response{}
}

//Cache related
func(this *ResponseBuilder) CachePublic() *ResponseBuilder{
    this.setCache("public")
    return this
}
func(this *ResponseBuilder) CachePrivate() *ResponseBuilder{
    this.setCache("private")
    return this
}
func(this *ResponseBuilder) CacheNoCache() *ResponseBuilder{
    this.setCache("no-cache")
    return this
}
func(this *ResponseBuilder) CacheNoStore() *ResponseBuilder{
    this.setCache("no-store")
    return this
}
func(this *ResponseBuilder) CacheNoTransform() *ResponseBuilder{
    this.setCache("no-transform")
    return this
}
func(this *ResponseBuilder) CacheMustReval() *ResponseBuilder{
    this.setCache("must-revalidate")
    return this
}
func(this *ResponseBuilder) CacheProxyReval() *ResponseBuilder{
    this.setCache("proxy-revalidate")
    return this
}
func(this *ResponseBuilder) CacheMaxAge(seconds int) *ResponseBuilder{
    this.setCache("max-age = "+string(seconds))
    return this
}
func(this *ResponseBuilder) CacheSMaxAge(seconds int) *ResponseBuilder{
    this.setCache("s-maxage = "+string(seconds))
    return this
}
func(this *ResponseBuilder) CacheClearAllOptions() *ResponseBuilder{
    this.writer().Header().Del("Cache-control")
    return this
}
func(this *ResponseBuilder) ConnectionKeepAlive(seconds int) *ResponseBuilder{
    this.writer().Header().Set("Connection", "keep-alive")
    return this
}
func(this *ResponseBuilder) ConnectionClose(seconds int) *ResponseBuilder{
    this.writer().Header().Set("Connection", "close")
    return this
}

func(this *ResponseBuilder) setCache(option string){
    this.writer().Header().Add("Cache-control", option)
}
