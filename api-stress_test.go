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

func testStress(t *testing.T) {
	loop1(t)
	loop2(t)
}

func loop1(t *testing.T) {
	//loop1 EndPoint `method:"DELETE" 	path:"/loop1/{Bool:bool}/mix1/{Int:int}"`
	//*******************************

	ch := make(chan int)
	fn := func(ch chan int, count int) {
		rb, _ := NewRequestBuilder(RootPath + "stress-service/loop1/true/mix1/5" + xrefStr)
		rb.AddCookie(cook)
		res, _ := rb.Delete()
		AssertEqual(res.StatusCode, 200, "Delete ResponseCode", t)
		ch <- count
	}

	runs := 200
	total := 0

	for i := 0; i < runs; i++ {
		go fn(ch, i)
		total = total + i
	}

	for i := 0; i < runs; i++ {
		total = total - <-ch
	}

	AssertEqual(total, 0, "testStress --> loop1", t)

}

func loop2(t *testing.T) {
	//loop2() EndPoint `method:"OPTIONS" path:"/loop2/{Bool:bool}/mix1/{Int:int}"`
	//*******************************

	ch := make(chan int)
	fn := func(ch chan int, count int) {
		strArr := make([]string, 0)
		rb, _ := NewRequestBuilder(RootPath + "stress-service/loop2/true/mix1/5" + xrefStr)
		rb.AddCookie(cook)
		res, _ := rb.Options(&strArr)
		AssertEqual(res.StatusCode, 200, "testStress --> loop2: Options ResponseCode", t)
		AssertEqual(len(strArr), 3, "testStress --> loop2: Options - slice length", t)
		if len(strArr) == 3 {
			AssertEqual(strArr[0], GET, "testStress --> loop2: Options", t)
			AssertEqual(strArr[1], HEAD, "testStress --> loop2: Options", t)
			AssertEqual(strArr[2], POST, "testStress --> loop2: Options", t)
		}
		ch <- count
	}

	runs := 200
	total := 0

	for i := 0; i < runs; i++ {
		go fn(ch, i)
		total = total + i
	}

	for i := 0; i < runs; i++ {
		total = total - <-ch
	}

	AssertEqual(total, 0, "testStress --> loop2", t)

}
