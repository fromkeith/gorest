// client.go
package main

import (
	"time"
)



func StartClient() {
	rb,er:=gorest.NewRequestBuilder()
	if er!=nil{
		println("Error, ",er.String())
	}
	u:=new(User)
	
	startSecs:=time.Nanoseconds()
	println("Starting loop")
	for i:=0;i<1000;i++{
		rb.Get(u,"http://localhost:8787/orders-service/users/3")
	}
	
	endTime:= time.Nanoseconds()
	
	//println("Process took: ", seconds := float64(endTime-startSecs)/1e9)
	
	
	
}

func TestEntity(){

	rb,er:=gorest.NewRequestBuilder()
	if er!=nil{
		println("Error, ",er.String())
	}
	u:=new(User)
	
	rb.Get(u,"http://localhost:8787/orders-service/users/3")
	println("Result: ",u.FirstName,u.LastName)
	
	items:=make([]Item,0)
	rb.Get(&items,"http://localhost:8787/orders-service/items")
	
	for _,i:= range items{
		println(i.Name," stock: ",i.AvailableStock)
	}
	
	
}
