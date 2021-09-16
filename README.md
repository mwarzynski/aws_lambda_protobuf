AWS Lambda with protobuf and Go

---

Repository contains proof of concept for Go application running in AWS Lambda
which responds to clients using `protobuf` to prepare requests.

Protobuf is most commonly used with gRPC. However, usage of gRPC is not possible with AWS Lambda (as of 09.2021).
Therefore I tried to check in practice if we could use proto API contracts to talk with Lambda.

Architecture of the proof of concept is shown below:

![Architecture](https://raw.github.com/mwarzynski/aws_lambda_protobuf/master/infra/docs/architecture.svg)

