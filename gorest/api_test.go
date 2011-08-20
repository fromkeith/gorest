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
	"strconv"
	"http"
)

type User struct {
	Id        string
	FirstName string
	LastName  string
	Age       int
	Weight    float32
}
//Tests:
//	Path data types: string,int,bool,float32,float64 
//	Returned(GET) data types: all basic ones above plus string keyed maps and structs.
//  Posted (POST) data types: same as GET above

type Service struct {
	RestService `root:"/serv/" consumes:"application/json" produces:"application/json" realm:"testing"`

	getVarArgs       EndPoint `method:"GET" path:"/var/{...:int}" output:"string"`
	postVarArgs      EndPoint `method:"POST" path:"/var/{...:int}" postdata:"string"`
	getVarArgsString EndPoint `method:"GET" path:"/varstring/{...:string}" output:"string"`

	getString EndPoint `method:"GET" path:"/string/{Bool:bool}/{Int:int}?{flow:int}&{name:string}" output:"string" role:"admin"`

	getInteger     EndPoint `method:"GET" path:"/int/{Bool:bool}/int/yes/{Int:int}/for" output:"int"`
	getBool        EndPoint `method:"GET" path:"/bool/{Bool:bool}/{Int:int}" output:"bool"`
	getFloat       EndPoint `method:"GET" path:"/float/{Bool:bool}/{Int:int}" output:"float64"`
	getMapInt      EndPoint `method:"GET" path:"/mapint/{Bool:bool}/{Int:int}" output:"map[string]int"`
	getMapStruct   EndPoint `method:"GET" path:"/mapstruct/{Bool:bool}/{Int:int}" output:"map[string]User"`
	getArrayStruct EndPoint `method:"GET" path:"/arraystruct/{FName:string}/{Age:int}" output:"[]User"`

	postString      EndPoint `method:"POST" path:"/string/{Bool:bool}/{Int:int}" postdata:"string" `
	postInteger     EndPoint `method:"POST" path:"/int/{Bool:bool}/{Int:int}" postdata:"int" `
	postBool        EndPoint `method:"POST" path:"/bool/{Bool:bool}/{Int:int}" postdata:"bool" `
	postFloat       EndPoint `method:"POST" path:"/float/{Bool:bool}/{Int:int}" postdata:"float64" `
	postMapInt      EndPoint `method:"POST" path:"/mapint/{Bool:bool}/{Int:int}" postdata:"map[string]int" `
	postMapStruct   EndPoint `method:"POST" path:"/mapstruct/{Bool:bool}/{Int:int}" postdata:"map[string]User" `
	postArrayStruct EndPoint `method:"POST" path:"/arraystruct/{Bool:bool}/{Int:int}" postdata:"[]User"`
}

type Complex struct {
	Auth       string `Header:""`
	Pathy      int    `Path:"Bool"`
	Query      int    `Query:"flow"`
	CookieUser string `Cookie:"User"`
	CookiePass string `Cookie:"Pass"`
}

var idsInRealm = map[string]string
type StrArr []string
func TestingAuthorizer(id string, role string) (bool, bool) {
	if idsInRealm ==nil{
		idsInRealm =make(map[string][]string,0)
		idsInRealm["123445"] =[]string{"admin", "user", "updater"}
		idsInRealm["fox"] =[]string{ "user"}
		idsInRealm["tom"] =[]string{ "user", "updater"}
	}
	
	if roles,found:=idsInRealm[id];found{
		for _, r := range roles {
			if role == r {
				return true,true
			}
		} 
		return true,false
	}
	
	return false,false
}




func (serv Service) GetVarArgs(v ...int) string {
	str := "Start"
	for _, i := range v {
		str += strconv.Itoa(i)
	}
	return str + "End"
}
func (serv Service) GetVarArgsString(v ...string) string {
	str := "Start"
	for _, i := range v {
		str += i
	}
	return str + "End"
}
func (serv Service) PostVarArgs(name string, varArgs ...int) {
	if name == "hello" && varArgs[0] == 5 && varArgs[1] == 24567 {
		serv.ResponseBuilder().SetResponseCode(200)
	} else {
		serv.ResponseBuilder().SetResponseCode(404)
	}

}
func (serv Service) GetString(Bool bool, Int int, Flow int, Name string) string {
	return "Hello" + strconv.Btoa(Bool) + strconv.Itoa(Int) + "/" + Name + strconv.Itoa(Flow)
}
func (serv Service) GetInteger(Bool bool, Int int) int {
	return Int - 5
}
func (serv Service) GetBool(Bool bool, Int int) bool {
	return Bool
}
func (serv Service) GetFloat(Bool bool, Int int) float64 {

	return 111.111 * float64(Int)
}
func (serv Service) GetMapInt(Bool bool, Int int) map[string]int {
	mp := make(map[string]int, 0)
	mp["One"] = 1
	mp["Two"] = 2
	mp["Three"] = 3
	return mp
}
func (serv Service) GetMapStruct(Bool bool, Int int) map[string]User {
	mp := make(map[string]User, 0)
	mp["One"] = User{"1", "David1", "Gueta1", 35, 123}
	mp["Two"] = User{"2", "David2", "Gueta2", 35, 123}
	mp["Three"] = User{"3", "David3", "Gueta3", 35, 123}
	return mp
}

func (serv Service) GetArrayStruct(FName string, Age int) []User {
	users := make([]User, 0)
	users = append(users, User{"user1", FName, "Soap", Age, 89.7})
	users = append(users, User{"user2", FName, "Soap2", Age, 89.7})
	return users
}


func (serv Service) PostString(posted string, Bool bool, Int int) {
	if posted == "Hello" && Bool && Int == 5 {
		serv.ResponseBuilder().SetResponseCode(200)
	} else {
		serv.ResponseBuilder().SetResponseCode(404)
	}
	println("posted:", posted)
}
func (serv Service) PostInteger(posted int, Bool bool, Int int) {
	if posted == 6 && Bool && Int == 5 {
		serv.ResponseBuilder().SetResponseCode(200)
	} else {
		serv.ResponseBuilder().SetResponseCode(404)
	}
	println("posted:", posted)
}
func (serv Service) PostBool(posted bool, Bool bool, Int int) {
	if !posted && Bool && Int == 5 {
		serv.ResponseBuilder().SetResponseCode(200)
	} else {
		serv.ResponseBuilder().SetResponseCode(404)
	}
	println("posted:", posted)
}
func (serv Service) PostFloat(posted float64, Bool bool, Int int) {
	if posted == 34.56788 && Bool && Int == 5 {
		serv.ResponseBuilder().SetResponseCode(200)
	} else {
		serv.ResponseBuilder().SetResponseCode(404)
	}
	println("posted:", posted)
}
func (serv Service) PostMapInt(posted map[string]int, Bool bool, Int int) {

	if posted["One"] == 111 && posted["Two"] == 222 && Bool && Int == 5 {
		serv.ResponseBuilder().SetResponseCode(200)
	} else {
		serv.ResponseBuilder().SetResponseCode(404)
	}
	println("posted map One:", posted["One"])
	println("posted map Two:", posted["Two"])
}
func (serv Service) PostMapStruct(posted map[string]User, Bool bool, Int int) {
	if posted["One"].FirstName == "David1" && posted["Two"].LastName == "Gueta2" && Bool && Int == 5 {
		serv.ResponseBuilder().SetResponseCode(200)
	} else {
		serv.ResponseBuilder().SetResponseCode(404)
	}
	println("posted map One:", posted["One"].FirstName, posted["One"].LastName, posted["One"].Id)
	println("posted map Two:", posted["Two"].FirstName, posted["Two"].LastName, posted["Two"].Id)
}
func (serv Service) PostArrayStruct(posted []User, Bool bool, Int int) {
	if posted[0].FirstName == "Joe" && posted[1].LastName == "Soap2" && Bool && Int == 5 {
		serv.ResponseBuilder().SetResponseCode(200)
	} else {
		serv.ResponseBuilder().SetResponseCode(404)
	}
	println("posted Array One:", posted[0].FirstName, posted[0].LastName, posted[0].Id)
	println("posted Array Two:", posted[1].FirstName, posted[1].LastName, posted[1].Id)

}

func TestInit(t *testing.T) {
	RegisterRealmAuthorizer("testing", TestingAuthorizer)
	RegisterService(new(Service))
	go ServeStandAlone(8787)

	cook := new(http.Cookie)
	cook.Name = "RestId"
	cook.Value = "12345"

	rb, _ := NewRequestBuilder()

	rb.AddCookie(cook)
	//GET string
	str := "Hell"
	res, _ := rb.Get(&str, "http://localhost:8787/serv/string/true/5?name=Nameed&flow=6")
	AssertEqual(res.StatusCode, 200, "Get string ResponseCode", t)
	AssertEqual(str, "Hellotrue5/Nameed6", "Get string", t)

	res, _ = rb.Get(&str, "http://localhost:8787/serv/string/true/5?name=Nameed")
	AssertEqual(res.StatusCode, 200, "Get string ResponseCode", t)
	AssertEqual(str, "Hellotrue5/Nameed0", "Get string", t)

	res, _ = rb.Get(&str, "http://localhost:8787/serv/string/true/5?flow=6")
	AssertEqual(res.StatusCode, 200, "Get string ResponseCode", t)
	AssertEqual(str, "Hellotrue5/6", "Get string", t)

	res, _ =rb.Get(&str, "http://localhost:8787/serv/string/true/5?flow=")
	AssertEqual(res.StatusCode, 200, "Get string ResponseCode", t)
	AssertEqual(str, "Hellotrue5/0", "Get string", t)

	res, _ =rb.Get(&str, "http://localhost:8787/serv/string/true/5?flow")
	AssertEqual(res.StatusCode, 200, "Get string ResponseCode", t)
	AssertEqual(str, "Hellotrue5/0", "Get string", t)

	res, _ =rb.Get(&str, "http://localhost:8787/serv/varstring/One/Two/Three")
	AssertEqual(res.StatusCode, 200, "Get string ResponseCode", t)
	AssertEqual(str, "StartOneTwoThreeEnd", "Get string", t)

	res, _ =rb.Get(&str, "http://localhost:8787/serv/var/1/2/3/4/5/6/7/8")
	AssertEqual(res.StatusCode, 200, "Get via var args ResponseCode", t)
	AssertEqual(str, "Start12345678End", "Get via var args", t)

	//GET Int
	inter := -2
	res, _ = rb.Get(&inter, "http://localhost:8787/serv/int/true/int/yes/2/for?name=Nameed&flow=6") //The query aurguments here just to be ignored
	AssertEqual(res.StatusCode, 200, "Get int ResponseCode", t)
	AssertEqual(inter, -3, "Get int", t)

	//GET Bool
	bl := true
	res, _ =rb.Get(&bl, "http://localhost:8787/serv/bool/false/2")
	AssertEqual(res.StatusCode, 200, "Get int ResponseCode", t)
	AssertEqual(bl, false, "Get Bool", t)

	//GET Float
	fl := 2.4
	res, _ = rb.Get(&fl, "http://localhost:8787/serv/float/false/2")
	AssertEqual(res.StatusCode, 200, "Get Float ResponseCode", t)
	AssertEqual(fl, 222.222, "Get Float", t)

	//GET Map Int
	mp := make(map[string]int)
	res, _ = rb.Get(&mp, "http://localhost:8787/serv/mapint/false/2")
	AssertEqual(res.StatusCode, 200, "Get Float ResponseCode", t)
	AssertEqual(mp["One"], 1, "Get Map Int", t)
	AssertEqual(mp["Two"], 2, "Get Map Int", t)

	//GET Map Int
	mpu := make(map[string]User)
	res, _ = rb.Get(&mpu, "http://localhost:8787/serv/mapstruct/false/2")
	AssertEqual(res.StatusCode, 200, "Get Map struct ResponseCode", t)
	AssertEqual(mpu["One"].Id, "1", "Get Map struct", t)
	AssertEqual(mpu["Two"].Id, "2", "Get Map struct", t)
	AssertEqual(mpu["Two"].FirstName, "David2", "Get Map struct", t)
	AssertEqual(mpu["Two"].LastName, "Gueta2", "Get Map struct", t)

	//GET Array Struct
	au := make([]User, 0)
	res, _ = rb.Get(&au, "http://localhost:8787/serv/arraystruct/Sandy/2")
	AssertEqual(res.StatusCode, 200, "Get Array struct ResponseCode", t)
	if res.StatusCode == 200{
		AssertEqual(au[0].Id, "user1", "Get Array Struct", t)
		AssertEqual(au[0].FirstName, "Sandy", "Get Array Struct", t)
	}
	

	//POST 

	res, _ = rb.Post("Hello", "http://localhost:8787/serv/string/true/5")
	AssertEqual(res.StatusCode, 200, "Post String", t)
	res, _ = rb.Post(6, "http://localhost:8787/serv/int/true/5")
	AssertEqual(res.StatusCode, 200, "Post Integer", t)
	res, _ = rb.Post(false, "http://localhost:8787/serv/bool/true/5")
	AssertEqual(res.StatusCode, 200, "Post Boolean", t)
	res, _ = rb.Post(34.56788, "http://localhost:8787/serv/float/true/5")
	AssertEqual(res.StatusCode, 200, "Post Float", t)

	//Post VarArgs
	res, _ = rb.Post("hello", "http://localhost:8787/serv/var/5/24567")
	AssertEqual(res.StatusCode, 200, "Post Var args", t)

	//POST Map Int
	mi := make(map[string]int, 0)
	mi["One"] = 111
	mi["Two"] = 222
	res, _ = rb.Post(mi, "http://localhost:8787/serv/mapint/true/5")
	AssertEqual(res.StatusCode, 200, "Post Integer Map", t)

	//POST Map Struct
	mu := make(map[string]User, 0)
	mu["One"] = User{"111", "David1", "Gueta1", 35, 123}
	mu["Two"] = User{"222", "David2", "Gueta2", 35, 123}
	res, _ = rb.Post(mu, "http://localhost:8787/serv/mapstruct/true/5")
	AssertEqual(res.StatusCode, 200, "Post Struct Map", t)

	//POST Array Struct
	users := make([]User, 0)
	users = append(users, User{"user1", "Joe", "Soap", 19, 89.7})
	users = append(users, User{"user2", "Jose", "Soap2", 15, 89.7})
	res, _ = rb.Post(users, "http://localhost:8787/serv/arraystruct/true/5")
	AssertEqual(res.StatusCode, 200, "Post Struct Array", t)

}

func TestServiceMeta(t *testing.T) {
	if meta, found := restManager.serviceTypes["gorest.googlecode.com/hg/gorest/Service"]; !found {
		t.Error("Service Not registered correctly")
	} else {
		AssertEqual(meta.consumesMime, "application/json", "Service consumesMime", t)
		AssertEqual(meta.producesMime, "application/json", "Service producesMime", t)
		AssertEqual(meta.root, "/serv/", "Service root", t)

	}

}
/*
func TestUsersByNameAndAge_Registration(t *testing.T){
	if ep,found:=restManager.endpoints[GET+":"+"serv/person/{FName:string}/{Age:int}"];!found{
		t.Error("Endpoint not registered.")
	}else{

		AssertEqual(ep.name,"usersByNameAndAge","endPoint name",t)
		AssertEqual(ep.requestMethod,GET,"endPoint method",t)

		AssertEqual(ep.signiture,"serv/person/{FName:string}/{Age:int}","endPoint signiture",t)
		AssertEqual(ep.root,"serv/person/","endPoint root",t)

		AssertEqual(ep.signitureLen,4,"endPoint signiture length",t)
		AssertEqual(ep.paramLen,2,"endPoint path param length",t)

		AssertEqual(ep.outputType,"User","method output",t)
		AssertEqual(ep.outputTypeIsArray,true,"method output array",t)

		AssertEqual(ep.parentTypeName,"gorest.googlecode.com/hg/gorest/Service","method output array",t)

	}
}


func TestGetUrl(t *testing.T){
	url:="/serv/person/Siya/444"
	if _,args,found:=getEndPointByUrl(GET,url);!found{
		t.Error("Fail Find: service endpoint from url:",url)
	}else{
		AssertEqual(args[0].parameter.name,"FName","Param Name",t)
		AssertEqual(args[0].parameter.typeName,"string","Param type",t)
		AssertEqual(args[0].data,"Siya","Param data",t)

		AssertEqual(args[1].parameter.name,"Age","Param Name",t)
		AssertEqual(args[1].parameter.typeName,"int","Param type",t)
		AssertEqual(args[1].data,"444","Param data",t)
	}
}*/
