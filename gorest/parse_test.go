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
	"testing"
)


type User struct{
    Id string
    FirstName string
    LastName string
    Age int
    Weight float32
}

type Service struct{
    RestService    "root=/serv/; consumes=application/json; produces=application/json"

    usersByNameAndAge EndPoint "method=GET; path=/person/{FName:string}/{Age:int}; output=[]User"
}

func(serv Service)    UsersByNameAndAge(FName string,Age int) []User{
    users:=make([]User,0)
    users=append(users,User{"user1",FName,"Soap",Age,89.7})
    users=append(users,User{"user2",FName,"Soap2",Age,89.7})
    return users
}


func TestServiceMetaData(t *testing.T){
    meta:= prepServiceMetaData("root=/serv/; consumes=application/json; produces=application/xml",new(Service))

    if meta.consumesMime != "application/json" {
       t.Error("Parsed incorrectly: 'consumesMime'")
    }
    if meta.producesMime != "application/xml" {
       t.Error("Parsed incorrectly: 'producesMime' ")
    }
    if meta.root != "/serv/" {
       t.Error("Parsed incorrectly: root ")
    }

}

func TestEndPointStruct(t *testing.T){
    meta:=makeEndPointStruct("method=GET; path=/person/{FName:string}/{Age:int}; output=[]User","/serv/")

    if meta.requestMethod != GET{
       t.Error("Parsed incorrectly: request method ")
    }
    if meta.root != "serv/person/"{
       t.Error("Parsed incorrectly: root ")
    }

    if meta.outputType == "User" && meta.outputTypeIsArray{
    }else{
       t.Error("Parsed incorrectly: return parameter ")
    }

    if meta.paramLen != 2{
       t.Error("Parsed incorrectly: parameter length ")
    }else if meta.params[0].name != "FName"  && meta.params[0].typeName != "User"{
       t.Error("Parsed incorrectly: path parameter names and types")
    }


}
