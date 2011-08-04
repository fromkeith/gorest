/**
 * Created by IntelliJ IDEA.
 * User: DlaminiSi
 * Date: 2011/06/09
 * Time: 7:22 PM
 * To change this template use File | Settings | File Templates.
 */
package main

import "json"
import "log"
import "os"
//import "fmt"
//import "strconv"

type Persona struct{
    Name string;
    Age int;

}
//
//func main(){
//     p:= new(Persona)
//     p.Name = "Siya"
//     p.Age = 29
//     marshall(p)
//}




func TestJson(){
     p:= new(Persona)
     p.Name = "Siya"
     p.Age = 29
     marshall(*p)
}

func marshall(p Persona){
   bytes,_:= json.Marshal(p)

    log.Println("Name: ",p.Name)
    log.Println("Age: ",p.Age)
//    log.Println("Bytes: ",bytes)

//    for _,b:=range bytes{
//        fmt.Println(strconv.Atob(b))
//    }
//    fmt.Println(json.)


    pers := new(Persona)
    json.Unmarshal(bytes,&pers)

    log.Println("Name: ",pers.Name)
    log.Println("Age: ",pers.Age)

    enc := json.NewEncoder(os.Stdout)
    enc.Encode(pers)

}
