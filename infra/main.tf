terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

# Configure the AWS Provider
provider "aws" {
  region = "us-east-2"
}

# Create a VPC
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "lambda_policy" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}


data "archive_file" "server" {
  type        = "zip"
  source_file = "./../function/cmd/server/server"
  output_path = "./server.zip"
}

resource "aws_lambda_function" "grpc" {
  filename      = data.archive_file.server.output_path
  function_name = "grpc-lambda"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "server"

  source_code_hash = filebase64sha256(data.archive_file.server.output_path)

  runtime = "go1.x"
}

resource "aws_apigatewayv2_api" "hello" {
  name          = "hello"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "hello" {
  api_id = aws_apigatewayv2_api.hello.id

  name        = "dev"
  auto_deploy = true
}

resource "aws_apigatewayv2_integration" "hello" {
  api_id = aws_apigatewayv2_api.hello.id

  integration_uri    = aws_lambda_function.grpc.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "hello-grpc" {
  api_id = aws_apigatewayv2_api.hello.id

  route_key = "POST /hello"
  target    = "integrations/${aws_apigatewayv2_integration.hello.id}"
}

resource "aws_lambda_permission" "api_gw" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.grpc.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.hello.execution_arn}/*/*"
}
