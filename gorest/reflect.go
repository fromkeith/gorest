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


import "reflect"
import "strings"
import "log"
import "strconv"
import "bytes"
import "http"

import "io"




const(
    ERROR_INVALID_INTERFACE="RegisterService(interface{}) takes a pointer to a struct that inherits from type RestService. Example usage: gorest.RegisterService(new(ServiceOne)) "
)


//Bootstrap functions below
//------------------------------------------------------------------------------------------


//Takes a value of a struct representing a service.
func RegisterService(h interface{}){

    if _,ok:= h.(GoRestService);!ok{
        panic(ERROR_INVALID_INTERFACE)
    }

    t:=reflect.TypeOf(h)

    if t.Kind() == reflect.Ptr{
        t = t.Elem()
    }else{
       panic(ERROR_INVALID_INTERFACE)
    }

    if t.Kind() == reflect.Struct{
        if field,found:=t.FieldByName("RestService");found{
            tag:=field.Tag
            meta:=prepServiceMetaData(tag,h)
            tFullName:=_manager().addType(t.PkgPath()+"/" + t.Name(),meta)
            for i:=0; i<t.NumField();i++{
                f:=t.Field(i)
                mapFieldsToMethods(t,f,tFullName,meta.root)
            }
        }
        return
    }

    panic(ERROR_INVALID_INTERFACE)
}


func mapFieldsToMethods(t reflect.Type, f reflect.StructField,typeFullName string,serviceRoot string){
    println("Type Fullname",typeFullName,f.Name)
   if f.Name != "RestService" && f.Type.Name() == "EndPoint"{   //TODO: Proper type checking, not by name
      println(f.Name,f.Type.Name())
      ep:=makeEndPointStruct(f.Tag,serviceRoot)
      ep.parentTypeName = typeFullName
      println("Parent type name now  set: ", typeFullName)


      var method reflect.Method
      methodName:= strings.ToUpper(f.Name[:1]) + f.Name[1:]
      println(methodName, ":", ep.requestMethod,ep.signiture,ep.outputType,ep.root,ep.paramLen)


      methFound:=false
      methodNumberInParent:=0
      for i:=0; i<t.NumMethod();i++{
         m:=t.Method(i)
         if methodName== m.Name{
            method = m //As long as the name is the same, we know we have found the method, since has no overloading
            methFound=true
            methodNumberInParent =i
            break
         }
      }


      { //Panic Checks
          if !methFound{
            panic("Method name not found. "+ panicMethNotFound(methFound ,ep ,t , f,methodName ))
          }
          if !isLegalForRequestType(method.Type,ep){
            panic("Parameter list not matching. "+panicMethNotFound(methFound ,ep ,t , f,methodName ))
          }
      }
      ep.methodNumberInParent = methodNumberInParent
      _manager().addEndPoint(ep)
      log.Println("Registerd service:",t.Name()," endpoint:",ep.requestMethod,ep.signiture)
   }
}

func isLegalForRequestType(methType reflect.Type, ep endPointStruct) (cool bool){
    cool =true

    numInputIgnore:=0
    numOut:=0

    switch ep.requestMethod{
        case POST,PUT: {
           numInputIgnore=2     //The first param is the struct, the second the posted object
           numOut=0
        }
        case GET:{
           numInputIgnore=1      //The first param is the struct
           numOut=1
        }
        case DELETE:{
           numInputIgnore=1      //The first param is the struct
           numOut=0
        }
    }

    if (methType.NumIn()-numInputIgnore) != ep.paramLen{
       cool=false
    }else if methType.NumOut()!=numOut{
       cool=false
    }else{
        //Check the first parameter type for POST and PUT
        if numInputIgnore ==2{
            if methType.In(1).Name() != ep.postdataType{
                cool=false
                return
            }
        }
        //Check the rest of the input param types
        for i:=numInputIgnore;i<methType.NumIn();i++{
            println(methType.In(i).Name())
            if methType.In(i).Name() != ep.params[i-numInputIgnore].typeName{

                cool=false
                break
            }
        }
        //Check output param type.
        if numOut==1{
            methVal:=methType.Out(0)
            if  ep.outputTypeIsArray{
               if  methVal.Kind() == reflect.Slice{
                    methVal=methVal.Elem()       //Only convert if it is mentioned as a slice in the tags, otherwise allow for failure panic
               }else{
                  cool=false
                  return
               }
            }

            if methVal.Name() != ep.outputType{
                cool=false
            }
        }
    }

    return
}

func panicMethNotFound(methFound bool,ep endPointStruct,t reflect.Type, f reflect.StructField,methodName string) string{

        var str string
        isArr:=""
        if ep.outputTypeIsArray{
            isArr="[]"
        }
        var suffix string = "("+isArr+ep.outputType+")# with one("+isArr+ep.outputType+") return parameter."
        if ep.requestMethod ==POST || ep.requestMethod ==PUT{
            str ="PostData "+ep.postdataType
            if ep.paramLen >0{
                str+=", "
            }

        }
        if ep.requestMethod ==POST || ep.requestMethod ==PUT || ep.requestMethod ==DELETE{
            suffix = "# with no return parameters."
        }
        for i:=0;i<ep.paramLen;i++{
            str+=ep.params[i].name + " " + ep.params[i].typeName
            if((i+1)<ep.paramLen){
                str+=","
            }
        }
        return "No matching Method found for EndPoint:["+f.Name+"],type:["+ep.requestMethod+"] . Expecting: #func(serv " + t.Name()+ ") "+ methodName + "("+str+")"+suffix
}





//Runtime functions below:
//-----------------------------------------------------------------------------------------------------------------


func prepareServe(context *Context,ep endPointStruct) ([]byte,restStatus) {
    servMeta:=_manager().getType(ep.parentTypeName)
    servInterface:=servMeta.template
    servVal:= reflect.ValueOf(servInterface).Elem()

    //Set the Context; the user can get the context from her services function param
    servVal.FieldByName("RestService").FieldByName("Context").Set(reflect.ValueOf(context))

    arrArgs:=make([]reflect.Value,0)

    //For POST and PUT, make and add the first "postdata" argument to the argument list
    if ep.requestMethod == POST || ep.requestMethod == PUT{
        //Make a new value from the methods postdata parameter type. The type validity should be checked at bootstrap time.
        //This parameter type is in position 1
        postdatVal:= reflect.New(servVal.Type().Method(ep.methodNumberInParent).Type.In(1))
        if postdatVal.Kind() == reflect.Ptr{
            postdatVal = postdatVal.Elem()
        }

        //Get postdata here
        //TODO: Also check if this is a multipart post and handle as required.
        buf := new(bytes.Buffer)
    	io.Copy(buf, context.request.Body)
        body := buf.String()

        println("This is the body of the post:",body)

        if v,state:=makeArg(ep.postdataType,body,postdatVal.Type(),servMeta.consumesMime);state.httpCode != http.StatusBadRequest{
           arrArgs=append(arrArgs,v)
        }else{
            return nil,state
        }
    }


    if len(context.args) == ep.paramLen{

      //Now add the rest of the arguments to the argument list and then call the method
      // GET and DELETE will only need these arguments, not the "postdata" one in their method calls
      for _,arg:=range context.args{
          if v,state:=makeArg(arg.parameter.typeName,arg.data,nil,"");state.httpCode != http.StatusBadRequest{
             arrArgs=append(arrArgs,v)
          }else{
            return nil,state
          }
      }

      //Now call the actual method with the data
      ret:=servVal.Method(ep.methodNumberInParent).Call(arrArgs)

      if len(ret)==1{  //This is when we have just called a GET
          //At the stage we should be ready to write the response to client

          if bytarr,err:=marshal(servMeta.producesMime,ret[0].Interface());err==nil{

             return bytarr,restStatus{http.StatusOK,""}

          }else{
            println("Error handling json...")
            //This is an internal error with the registered not being able to marshal internal structs
            return nil,restStatus{http.StatusInternalServerError,"Internal server error."}
          }
      }else{
            println("POST succsesfull, notthing to return")
            return nil,restStatus{http.StatusOK,""}
      }
    }

    //Just in case the whole civilization crashes and it falls thru to here. This shall never happen though... well tested
    log.Fatalln("There was a problem with request handing. Probably a bug, please report.") //Add client data, and send support alert
    return nil, restStatus{http.StatusInternalServerError,"Internal server error."}
}

func makeArg(typeName string, data string,template reflect.Type,mime string)(reflect.Value,restStatus){

    switch typeName{
        case "string": {
             return reflect.ValueOf(data) ,restStatus{http.StatusOK,""}
        }
        case "int": {
             if i,err:=strconv.Atoi(data);err==nil{
                return reflect.ValueOf(i),restStatus{http.StatusOK,""}
             }
        }
        case "bool": {
             if i,err:=strconv.Atob(data);err==nil{
                return reflect.ValueOf(i),restStatus{http.StatusOK,""}
             }
        }
        case "float32","float64":{
             if i,err:=strconv.Atof64(data);err==nil{
                return reflect.ValueOf(i),restStatus{http.StatusOK,""}
             }
        }
        default: {

             return callReqisteredUnMarhaller(data,mime,template)
        }
    }
    log.Println("Invalid user data inputed...",data)
    return reflect.ValueOf(nil),restStatus{http.StatusBadRequest,"Invalid user input."}
}

func callReqisteredUnMarhaller(data string,mime string,template reflect.Type)(reflect.Value,restStatus){

      i:= reflect.New(template).Interface()
      if err:=unMarshal(mime,[]byte(data),i);err!=nil{
        println("Error marhsaling string",err.String())
        return reflect.ValueOf(nil),restStatus{http.StatusBadRequest,"Error Unmarshalling data using "+mime+". Client sent incompetible data format in entity."}
      }
      return reflect.ValueOf(i).Elem(),restStatus{http.StatusOK,""}
}




