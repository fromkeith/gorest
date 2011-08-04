/**
 * Created by IntelliJ IDEA.
 * User: DlaminiSi
 * Date: 2011/07/26
 * Time: 7:22 PM
 * To change this template use File | Settings | File Templates.
 */
package main

import gorest "./gorest"

import "reflect"
//import "json"
//import "os"



func main(){
     RunReflectTest()
//     gorest.RunParseTest()

//    test()
//    TestJson()

}


func RunReflectTest(){
//    bytes,_:= json.Marshal(User{"Siyabonga",29})
//    println("bytes",bytes)
//    println("bytes",string(bytes))
//    gorest.TestServe("GET","/person/Siya/",string(bytes))

//    gorest.RegisterService(new(ServiceOne))
    gorest.RegisterService(new(ServiceOne))
    gorest.ServeStandAlone(8787)


}

type User struct{
    Name string
    Age int
}


type ServiceOne struct{
    gorest.RestService    "root=/serv/; consumes=application/json; produces=application/json"

    getMyUser gorest.EndPoint "method=GET; path=/person/{Name:string}/{Age:int}; output=string"
    postTheUser gorest.EndPoint "method=POST; path=/person/{Name:string}/{Age:int}; postdata=User"
    deleteDaUser gorest.EndPoint "method=DELETE; path=/person/{Name:string}/{Age:int};"
    put gorest.EndPoint "method=PUT; path=/person/{Name:string}/{Age:int}; postdata=User"

    getMyUsers gorest.EndPoint "method=GET; path=/person/{Name:string}; output=[]User"

}
func(serv ServiceOne)  GetMyUser(str string , num int) (string){
    println("Right inside Get ",str,num)

    rb:=serv.ResponseBuilder()
    rb.SetResponseCode(205).SetContentType(gorest.Application_Json).WriteAndContinue([]byte("Bye bye")).Send()

  return "Snacks"
}
func(serv ServiceOne)  PostTheUser(str User, Name string , Age int){
  println("Right inside Post",str.Name,str.Age,Name,Age)
}
func(serv ServiceOne)  Put(str User, Name string , Age int){
  println("Right inside Put",str.Name,str.Age,Name,Age)
}
func(serv ServiceOne)  DeleteDaUser(Name string , Age int){
  println("Right inside Delete",Name,Age)
}

func(serv ServiceOne)  GetMyUsers(str string ) ([]User){
    println("Right inside Get ",str)

    users:=make([]User,0)
    users=append(users,User{str+"1 Dlamini",0})
    users=append(users,User{str+"2 Dlamini",0})

  return   users
}




type My_struct struct{
    I int
    S string
    Sptr *int
}
func(my My_struct) Print(str string){
     for i:=1;i<=5;i++{
        println(i,str,*my.Sptr)
     }
}
func test(){
//    b:=5
    a := new(My_struct)//My_struct{1, "alpha",&b}
    inter(a)
}
func inter(a interface{}){


    sVal := reflect.ValueOf(a)

    if sVal.Kind() == reflect.Ptr{
       sVal = sVal.Elem()
       println("Setting to Elem",sVal.Kind())
    }

    if sVal.Kind() == reflect.Struct{
        println("struct val")
        i:=sVal.Field(0).Int()
        println(i)
        sVal.Field(0).SetInt(5)
        println(sVal.Field(0).Int())

        c:=7
//        reflect.Indirect(sVal.Field(2)).SetInt(6)

        sVal.Field(2).Set(reflect.ValueOf(&c))

        strVal:=reflect.ValueOf("Hello")
        arr:=make([]reflect.Value,0)
        arr=append(arr,strVal)
        sVal.Method(0).Call(arr)
    }

}

