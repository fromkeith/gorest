/**
 * Created by IntelliJ IDEA.
 * User: DlaminiSi
 * Date: 2011/06/09
 * Time: 7:22 PM
 * To change this template use File | Settings | File Templates.
 */
package main

import (
        rest "./rest"
        "fmt"
        "log"
        "json"

        )

//func main(){
////    str := "百度一下，你就知道"
////
//
//     rest.RegisterEndPoint("/services/pdf/rename/{oldName}/{newName}/{oldName1}/{newName1}",s1)
//     rest.RegisterEndPoint("/services/{newName1}",s2)
//     rest.RegisterEndPoint("/services/",s3)
//     rest.ServeStandAlone(8080)
//
//
//
//}


func s1(c *rest.Context){

     fmt.Fprintf(c.Writer(), "Hi there, I love %s!","this")
     log.Println("Serving..............Me Hurray!!!!!!!!!!!")
}

func s2(c *rest.Context){
     p:= new(Persona)
     p.Name = "Siya"
     p.Age = 29
//     bytes,_:= json.Marshal(p)


    enc := json.NewEncoder(c.Writer())
    enc.Encode(p)
//
//      log.Println("Serving: ", "2")
      fmt.Fprintf(c.Writer(),"Serving: ", "Hello")
}
func s3(c *rest.Context){
      log.Println("Serving: ", "3")
//      for c.
      fmt.Fprintf(c.Writer(),"Serving: ", "3")
}








