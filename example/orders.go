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


package main

import (
	"gorest.googlecode.com/hg/gorest"
	"strconv"
	)

//Use a rest client to add orders:
//Order Json Example:  {"Id":0,"ItemId":0,"UserId":0,"Amount":0,"Discount":0,"Cancelled":false}

func main(){
    gorest.RegisterService(new(OrderService))
    go gorest.ServeStandAlone(8787)
	TestEntity()
	
}

//************************Define Service***************************

type OrderService struct {
	//Service level config
	gorest.RestService `root:"/orders-service/" consumes:"application/json" produces:"application/json"`

	//End-Point level configs: Field names must be the same as the corresponding method names,
	// but not-exported (starts with lowercase)

	userDetails gorest.EndPoint `method:"GET" path:"/users/{Id:int}" output:"User"`
	listItems   gorest.EndPoint `method:"GET" path:"/items/" output:"[]Item"`
	addItem     gorest.EndPoint `method:"POST" path:"/items/" postdata:"Item"`

	//On real app for placeOrder below, the POST URL would probably be just /orders/, this is just to
	// demo the ability of mixing post-data parameters with URL mapped parameters.
	placeOrder  gorest.EndPoint `method:"POST" path:"/orders/new/{UserId:int}/{RequestDiscount:bool}/" postdata:"Order"`
	viewOrder   gorest.EndPoint `method:"GET" path:"/orders/{OrderId:int}" output:"Order"`
	//viewOrders   gorest.EndPoint `method:"GET" path:"/orders/{OrderId:int}" output:"map[string]Order"`
	deleteOrder gorest.EndPoint `method:"DELETE" path:"/orders/{OrderId:int}"`
}

//Handler Methods: Method names must be the same as in config, but exported (starts with uppercase)



func (serv OrderService) UserDetails(Id int) (u User) {
	if user, found := userStore[Id]; found {
		u = user
		return
	}
	serv.ResponseBuilder().SetResponseCode(404).Overide(true) //Overide causes the entity returned by the method to be ignored. Other wise it would send back zeroed object
	return
}

func (serv OrderService) ListItems() []Item {
	serv.ResponseBuilder().CacheMaxAge(60 * 60 * 24) //List cacheable for a day. More work to come on this, Etag, etc
	return itemStore
}

func (serv OrderService) AddItem(i Item) {

	for _, item := range itemStore {
		if item.Id == i.Id {
			item = i
			serv.ResponseBuilder().SetResponseCode(200) //Updated http 200, or you could just return without setting this. 200 is the default for POST
			return
		}
	}

	//Item Id not in database, so create new
	i.Id = len(itemStore)
	itemStore = append(itemStore, i)

	serv.ResponseBuilder().Created("http://localhost:8787/orders-service/items/" + string(i.Id)) //Created, http 201
}

//On the method parameters, the posted data(http-entity) is always first, followed by the URL mapped parameters
func (serv OrderService) PlaceOrder(order Order, UserId int, AskForDiscount bool) {
	order.Id = len(orderStore)

	if user, found := userStore[UserId]; found {
		if item, exists := findItem(order.ItemId); exists {
			itemStore[item.Id].AvailableStock--

			if AskForDiscount && order.Amount > 5 {
				order.Discount = 2.5
			}
			order.Id = len(orderStore)
			order.UserId = UserId
			order.Cancelled = false
			orderStore = append(orderStore, order)
			user.OrderIds = append(user.OrderIds, order.Id)

			userStore[user.Id] = user

			serv.ResponseBuilder().SetResponseCode(201).Location("http://localhost:8787/orders-service/orders/" + string(order.Id)) //Created
			return

		} else {
			serv.ResponseBuilder().SetResponseCode(404).WriteAndOveride([]byte("Item not found")) //You can still manually place an entity on the response, even on a POST
			return
		}
	}

	serv.ResponseBuilder().SetResponseCode(404).WriteAndOveride([]byte("User not found"))
	return
}
func (serv OrderService) ViewOrder(id int) (retOrder Order) {
	for _, order := range orderStore {
		if order.Id == id {
			retOrder = order
			return
		}
	}
	serv.ResponseBuilder().SetResponseCode(404).Overide(true)
	return
}
func (serv OrderService) DeleteOrder(id int) {
	for pos, order := range orderStore {
		if order.Id == id {
			order.Cancelled = true
			orderStore[pos] = order
			return //Default http code for DELETE is 200
		}
	}
	serv.ResponseBuilder().SetResponseCode(404).Overide(true)
	return
}

//*********************************End of service******************************************************


type Item struct {
	Id             int
	Name           string
	AvailableStock int
	Price          float32
}

type Order struct {
	Id        int
	ItemId    int
	UserId    int
	Amount    int
	Discount  float32
	Cancelled bool
}

type User struct {
	Id        int
	FirstName string
	LastName  string
	OrderIds  []int
}

var (
	itemStore  []Item
	orderStore []Order
	userStore  map[int]User
)

func init() {
	itemStore = make([]Item, 0)
	orderStore = make([]Order, 0)
	userStore = make(map[int]User, 0)

	initUsers()
	initItems()
}

func initUsers() {
	for i := 1; i <= 10; i++ {
		userStore[i] = User{Id: i,
			FirstName: "Username" + strconv.Itoa(i),
			LastName:  "Lastname" + strconv.Itoa(i)}
	}
}

func initItems() {
	for i := 1; i <= 10; i++ {
		itemStore = append(itemStore, Item{i, "Item: " + strconv.Itoa(i), i + 5, 89.5})
	}
}


func findItem(id int) (item Item, found bool) {
	for _, i := range itemStore {
		if i.Id == id {
			item, found = i, true
			return
		}
	}
	return
}
