AWS Lambda with protobuf and Go

---

Repository contains proof of concept for Go app running in AWS.
which responds to clients using `protobuf` to prepare requests.

Protobuf is most commonly used with gRPC. However, usage of gRPC isn't possible with AWS Lambda as of 09/2021.
Therefore this code checks in practice if it's possible to use proto API contracts to talk with Lambda.

Architecture of the proof of concept is as below:

![Architecture](https://raw.github.com/mwarzynski/aws_lambda_protobuf/master/infra/docs/architecture.svg)
