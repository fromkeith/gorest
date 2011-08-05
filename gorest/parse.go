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

import(
    "strings"
    "log"
)




type argumentData struct{
    parameter param
    data string
}
type param struct{
    positionInPath int
    name string
    typeName string
}

var ALLOWED_PAR_TYPES= []string{"string","int","bool","float32","float64"}


func prepServiceMetaData(tag string,i interface{}) serviceMetaData{
    md:=new(serviceMetaData)
    tag= strings.Trim(tag," ")
    tag= strings.Trim(tag,";")
    for _,str:= range strings.Split(tag,";",-1){
         str = strings.Trim(str," ")
         if strings.HasPrefix(str,"root="){
             name:= str[strings.Index(str,"=")+1:]
             md.root =name
         }else if strings.HasPrefix(str,"consumes="){
             name:= str[strings.Index(str,"=")+1:]
             md.consumesMime =name
         }else if strings.HasPrefix(str,"produces="){
             name:= str[strings.Index(str,"=")+1:]
             md.producesMime =name
         }else{
            panic("Unknown annotaion: "+str+"; in service declaration. Allowed types {root,consumes,produces}")
         }
    }
    md.template = i
    return *md
}

func makeEndPointStruct(tags string,serviceRoot string) endPointStruct{

    ms:=new(endPointStruct)
    tags= strings.Trim(tags," ")
    tags= strings.Trim(tags,";")

    for _,str:= range strings.Split(tags,";",-1){
       str = strings.Trim(str," ")
       if strings.HasPrefix(str,"method="){
            name:= str[strings.Index(str,"=")+1:]
            if name=="GET"{
                ms.requestMethod =GET
            }else if name=="POST"{
                ms.requestMethod =POST
            }else if name=="PUT"{
                ms.requestMethod =PUT
            }else if name=="DELETE"{
                ms.requestMethod =DELETE
            }else if name=="OPTIONS"{
                ms.requestMethod =OPTIONS
            }else {
                panic("Unknown method type:["+name+"] in endpoint declaration. Allowed types {GET,POST,PUT,DELETE,OPTIONS}")
            }
       }else if strings.HasPrefix(str,"path=") {
             serviceRoot = strings.TrimRight(serviceRoot,"/")
             name:= str[strings.Index(str,"=")+1:]
             ms.signiture = serviceRoot+ "/" + strings.TrimLeft(name,"/")
       }else if strings.HasPrefix(str,"output=") {
             name:= str[strings.Index(str,"=")+1:]
             ms.outputType = name
             if strings.HasPrefix(name,"[]"){  //Check for slice/array/list types. We only handle these on output!!!
                ms.outputTypeIsArray =true
                ms.outputType = ms.outputType[2:]
             }
       }else if strings.HasPrefix(str,"input=") {
             name:= str[strings.Index(str,"=")+1:]
             ms.inputMime = name
       }else if strings.HasPrefix(str,"postdata=") {
             name:= str[strings.Index(str,"=")+1:]
             ms.postdataType = name
       }else {
            panic("Unknown annotaion: "+str+"; in service declaration. Allowed types {method,path,output,input}")
       }
    }
    parseParams(ms)
    return *ms
}
func parseParams(e *endPointStruct){
    e.signiture=strings.Trim(e.signiture,"/")
    e.params= make(map[int]param,0)

    i:= strings.Index(e.signiture,"{")
    if i>0{
        e.root = e.signiture[0:i]
    }else{
        e.root = e.signiture
    }

    //Extract path parameters
    for pos,str1:= range strings.Split(e.signiture,"/",-1){
        e.signitureLen++
        if strings.HasPrefix(str1,"{") && strings.HasSuffix(str1,"}"){

           temp:=strings.Trim(str1,"{}")
           ind:=0
           if  ind=strings.Index(temp,":"); ind==-1{
             panic("Please ensure that parameter names("+temp+") have associated types in REST path: "+e.signiture)
           }
           parName:=temp[:ind]
           typeName:=temp[ind+1:]

           if !isAllowedParamType(typeName){
              panic("Type "+typeName+" is not allowed for path-parameters in REST path: "+e.signiture)
           }

           for _,par:=range e.params{
             if par.name==parName{
               panic("Duplicate field name("+parName+") in REST path: "+e.signiture )
             }
           }

           e.params[e.paramLen] = param{pos,parName,typeName}
           e.paramLen++
        }
    }

    if ep,there:=_manager().endpoints[e.signiture];there && ep.requestMethod == e.requestMethod{
        panic("Endpoint already registered: "+e.signiture)
    }

    for _,ep:= range _manager().endpoints{
       if ep.root == e.root && ep.signitureLen == e.signitureLen && ep.requestMethod == e.requestMethod{

        panic("Can not register two endpoints with same request-method("+ep.requestMethod+"), same root and same amount of parameters: "+e.signiture)
       }
    }
}

func isAllowedParamType(typeName string)bool{
    for _,s:=range ALLOWED_PAR_TYPES{
        if s == strings.ToLower(typeName){
            return true
        }
    }
    return false
}



func getEndPointByUrl(method string,url string) (endPointStruct,[]argumentData,bool){
    println("Looking for "+method+" url endpoint: ",url)
    url=strings.Trim(url,"/")
    totalParts:= strings.Count(url,"/")
    totalParts++

    log.Println("Endpoints: ",len(_manager().endpoints))
    epRet:=new(endPointStruct)
    adarr:=make([]argumentData,0)
    ok:=false
    for _,ep:= range _manager().endpoints{
       if strings.Contains(url,ep.root) && ep.signitureLen== totalParts && ep.requestMethod ==method{   //TODO: Make sure it starts with
           log.Println("End point found: ",ep.requestMethod,ep.root,ep.signiture,ep.signitureLen,url,ep.paramLen)
           for _,par:=range ep.params{
                for upos,str1:= range strings.Split(url,"/",-1){
                    if par.positionInPath==upos{
                        adarr=append(adarr,argumentData{par,strings.Trim(str1," ")})
                    }
                }
           }
           ok =true
           return ep,adarr,ok
       }
    }
    return  *epRet,adarr,ok
}
