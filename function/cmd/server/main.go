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
func lambdaHandler(request events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	if request.Path == "/health" {
		return events.ALBTargetGroupResponse{
			Body:              request.Body,
			StatusCode:        http.StatusOK,
			StatusDescription: http.StatusText(http.StatusOK),
			IsBase64Encoded:   false,
			Headers:           map[string]string{},
		}, nil
	}

	if request.HTTPMethod != "POST" {
		return events.ALBTargetGroupResponse{
			StatusCode:        http.StatusNoContent,
			StatusDescription: http.StatusText(http.StatusNoContent),
			IsBase64Encoded:   false,
			Headers:           map[string]string{},
		}, nil
	}

	if !request.IsBase64Encoded {
		return events.ALBTargetGroupResponse{
			Body:              fmt.Sprintf("expected base64 encoded body, got: %s", request.Body),
			StatusCode:        http.StatusUnprocessableEntity,
			StatusDescription: http.StatusText(http.StatusUnprocessableEntity),
			IsBase64Encoded:   false,
			Headers:           map[string]string{},
		}, nil
	}
	contentType := request.Headers["content-type"]
	if contentType != "application/protobuf" {
		return events.ALBTargetGroupResponse{
			Body:              fmt.Sprintf("expected 'application/proto' content type, got: %s", contentType),
			StatusCode:        http.StatusUnprocessableEntity,
			StatusDescription: http.StatusText(http.StatusUnprocessableEntity),
			IsBase64Encoded:   false,
			Headers:           map[string]string{},
		}, nil
	}

	reqBody, err := base64.StdEncoding.DecodeString(request.Body)
	if err != nil {
		return events.ALBTargetGroupResponse{
			Body:              fmt.Sprintf("isBase64Encoded = true, but body is not valid base64 encoding: %s", request.Body),
			StatusCode:        http.StatusBadRequest,
			StatusDescription: http.StatusText(http.StatusBadRequest),
			IsBase64Encoded:   false,
			Headers:           map[string]string{},
		}, nil
	}

	var req hello.HelloRequest
	if err := proto.Unmarshal(reqBody, &req); err != nil {
		return events.ALBTargetGroupResponse{
			Body:              fmt.Sprintf("request body is not a valid json for `HelloRequest`, got: %s", request.Body),
			StatusCode:        http.StatusBadRequest,
			StatusDescription: http.StatusText(http.StatusBadRequest),
			IsBase64Encoded:   false,
			Headers:           map[string]string{},
		}, nil
	}

	resp := HandleSayHello(&req)

	b, err := proto.Marshal(resp)
	if err != nil {
		return events.ALBTargetGroupResponse{
			Body:              fmt.Sprintf("proto marshal of response error: %s", err.Error()),
			StatusCode:        http.StatusInternalServerError,
			StatusDescription: http.StatusText(http.StatusInternalServerError),
			IsBase64Encoded:   false,
			Headers:           map[string]string{},
		}, err
	}

	return events.ALBTargetGroupResponse{
		Body:              base64.StdEncoding.EncodeToString(b),
		StatusCode:        200,
		StatusDescription: "200 OK",
		IsBase64Encoded:   true,
		Headers:           map[string]string{"Content-Type": "application/proto"},
	}, nil
}

func main() {
	lambda.Start(lambdaHandler)
}
