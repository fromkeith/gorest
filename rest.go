/**
 * Created by IntelliJ IDEA.
 * User: DlaminiSi
 * Date: 2011/06/09
 * Time: 5:45 PM
 * To change this template use File | Settings | File Templates.
 */
package rest

import (
    "http"
    "log"
    str "strings"
    "strconv"
)


type Context struct{
    writer http.ResponseWriter
    request *http.Request
    params map[string]string
}

type EndPoint struct{

    signiture string
	root string
	handler func(*Context);
    params map[string]int

    signitureLen int
    paramLen int
}



type goRest struct{
    endPoints   map[string]EndPoint
}

var restState goRest

func init(){
     restState.endPoints = make(map[string]EndPoint,0)
}





func RegisterEndPoint(signiture string, handler func(*Context)){

    e:= createEndPoint(signiture,handler)

    if _,there:=restState.endPoints[signiture];there{
        log.Panic("Endpoint already registered: "+signiture)
    }

    for _,ep:= range restState.endPoints{
       if ep.root == e.root && ep.signitureLen == e.signitureLen{

        log.Panic("Can not register two endpoints with same root and same amount of parameters: "+signiture)
       }
    }

    log.Println("Len of :", len(restState.endPoints))
    restState.endPoints[signiture] =  e


}

func createEndPoint(signiture string, handler func(*Context)) (EndPoint){
    signiture=str.Trim(signiture,"/")
    log.Println("Trimmed :",signiture)
    e := &EndPoint{signiture,"",handler,make(map[string]int,5),0,0}

    i:= str.Index(signiture,"{")
    if i>0{
        e.root = signiture[0:i]
    }else{
        e.root = signiture
    }

    //Extract parameters
    for pos,str1:= range str.Split(signiture,"/",-1){
        e.signitureLen++
        if str.HasPrefix(str1,"{") && str.HasSuffix(str1,"}"){
           e.paramLen++
           temp:=str.Trim(str1,"{}")
           if _,there:=e.params[temp];there{
             panic("Duplicate field name in REST path: "+signiture + "; field: "+temp)
           }
           e.params[temp] = pos
        }
    }

    return *e;
}


func getEndPointByUrl(url string) (*EndPoint,bool){
    url=str.Trim(url,"/")
    totalParts:= str.Count(url,"/")
    totalParts++

    log.Println("Parts: ",totalParts)

    for _,ep:= range restState.endPoints{
       log.Println("ROOT: ",ep.root,ep.signitureLen)
       log.Println("PATH: ",url,totalParts)
       if str.Contains(url,ep.root) && ep.signitureLen== totalParts{   //TODO: Make sure it starts with
           log.Println("End point found: ",ep.signiture)
           return &ep,true
       }
    }
    return nil,false
}

func mainHandler(w http.ResponseWriter, r *http.Request){

    log.Println("Serving URL : ",r.RawURL)

    if a,ok:= getEndPointByUrl(r.RawURL);ok{
        c:= new(Context)
        c.writer = w
        c.request = r
        c.params =  make(map[string]string)

        a.handler(c)
        log.Println("Serving with endpoint: ", a.signiture,ok)
    }else{
        log.Println("Could not find endpoint.")
        http.NotFound(w,r)
    }
}



func ServeStandAlone(port int){
    http.HandleFunc("/", mainHandler)
    http.ListenAndServe(":"+strconv.Itoa(port), nil)
}


func (c *Context) Writer() (http.ResponseWriter){
    return c.writer
}
func (c *Context) Request() (*http.Request){
    return c.request
}

func (c *EndPoint) Root() (string){
    return c.root;
}

func (c *EndPoint) SignitureLen() (int){
    return c.signitureLen;
}

func (c *EndPoint) ParamLen() (int){
    return c.paramLen;
}

func (c *EndPoint) ParamDefs() (map[string]int){
    return c.params;
}




