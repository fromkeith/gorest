/**
 * Created by IntelliJ IDEA.
 * User: DlaminiSi
 * Date: 2011/07/26
 * Time: 5:47 PM
 * To change this template use File | Settings | File Templates.
 */
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
