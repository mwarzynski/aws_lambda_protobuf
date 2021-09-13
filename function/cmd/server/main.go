package main

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"google.golang.org/protobuf/proto"

	"mwarzynski/aws-grpc-lambda/api/hello"
)

func HandleSayHello(req *hello.HelloRequest) *hello.HelloReply {
	return &hello.HelloReply{
		Message: fmt.Sprintf("Hello, %s. Golang, protobuf and AWS Lambda works!!1", req.Name),
	}
}

// lambdaHandler is our magic handler which understands proto language.
func lambdaHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !request.IsBase64Encoded {
		return events.APIGatewayProxyResponse{
			Body:       "body is not encoded with base64",
			StatusCode: http.StatusUnprocessableEntity,
		}, nil
	}
	if request.Headers["Content-Type"] != "application/protobuf" {
		return events.APIGatewayProxyResponse{
			Body:       "expected 'application/protobuf' content type",
			StatusCode: http.StatusUnprocessableEntity,
		}, nil
	}
	reqRaw, err := base64.StdEncoding.DecodeString(request.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("body is not encoded with base64 even though IsBase64Encoded = true, %s", request.Body),
			StatusCode: http.StatusBadRequest,
		}, nil
	}
	var req hello.HelloRequest
	if err := proto.Unmarshal(reqRaw, &req); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("request body is not a valid proto for `HelloRequest`, got: %s", request.Body),
			StatusCode: http.StatusUnprocessableEntity,
		}, nil
	}

	resp := HandleSayHello(&req)

	b, err := proto.Marshal(resp)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	return events.APIGatewayProxyResponse{
		Body:       string(b),
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(lambdaHandler)
}
