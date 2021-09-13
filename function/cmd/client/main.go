/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"net/http"

	"google.golang.org/protobuf/proto"

	"mwarzynski/aws-grpc-lambda/api/hello"
)

func main() {
	address := flag.String("url", "http://tf-hello-lambda-800097377.us-east-2.elb.amazonaws.com/hello", "")
	flag.Parse()
	println(*address)

	req := hello.HelloRequest{
		Name: "Mateusz",
	}

	b, err := proto.Marshal(&req)
	if err != nil {
		println("invalid message")
		println(err.Error())
		return
	}

	resp, err := http.Post(*address, "application/protobuf", bytes.NewReader([]byte(b)))
	if err != nil {
		println("making post request err")
		println(err.Error())
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var reply hello.HelloReply
	if err := proto.Unmarshal(body, &reply); err != nil {
		println("unmarshal")
		println(string(body))
		println(err.Error())
		return
	}

	println(reply.Message)
}
