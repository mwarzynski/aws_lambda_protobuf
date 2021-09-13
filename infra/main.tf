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
resource "aws_vpc" "main" {
  cidr_block       = "10.0.0.0/16"
  instance_tenancy = "default"
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

resource "aws_iam_role_policy_attachment" "AWSLambdaVPCAccessExecutionRole" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
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

  vpc_config {
    subnet_ids         = [aws_subnet.main1.id, aws_subnet.main2.id]
    security_group_ids = [aws_security_group.allow_everything.id]
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.main.id
}

resource "aws_route" "internet" {
  route_table_id         = aws_vpc.main.default_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.gw.id
}

resource "aws_route_table_association" "rta-main1" {
  route_table_id = aws_vpc.main.default_route_table_id
  subnet_id      = aws_subnet.main1.id
}

resource "aws_route_table_association" "rta-main2" {
  route_table_id = aws_vpc.main.default_route_table_id
  subnet_id      = aws_subnet.main2.id
}

resource "aws_subnet" "main1" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-2a"
}

resource "aws_subnet" "main2" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-east-2b"
}

resource "aws_security_group" "allow_everything" {
  name        = "allow-everything"
  description = "Allow everything"
  vpc_id      = aws_vpc.main.id

  ingress = [
    {
      description      = "allow everything in"
      from_port        = 0
      to_port          = 0
      protocol         = "-1"
      cidr_blocks      = ["0.0.0.0/0"]
      ipv6_cidr_blocks = ["::/0"]
      self             = null,
      prefix_list_ids  = []
      security_groups  = null
    }
  ]

  egress = [
    {
      description      = "allow everything out"
      from_port        = 0
      to_port          = 0
      protocol         = "-1"
      cidr_blocks      = ["0.0.0.0/0"]
      ipv6_cidr_blocks = ["::/0"]
      self             = null,
      prefix_list_ids  = []
      security_groups  = null
    }
  ]
}

resource "aws_lb" "alb" {

  name               = "tf-hello-lambda"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.allow_everything.id]
  subnets            = [aws_subnet.main1.id, aws_subnet.main2.id]

  enable_deletion_protection = false

  depends_on = [
    aws_internet_gateway.gw
  ]
}


resource "aws_lb_listener" "lambda" {
  load_balancer_arn = aws_lb.alb.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "Hello World!"
      status_code  = "200"
    }
  }
}

resource "aws_lb_listener_rule" "static" {
  listener_arn = aws_lb_listener.lambda.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.arn
  }

  condition {
    path_pattern {
      values = ["/hello*"]
    }
  }
}


resource "aws_lb_target_group" "main" {
  name        = "hello-lambda-tg"
  target_type = "lambda"

  health_check {
    healthy_threshold   = 2
    interval            = 60
    timeout             = 5
    unhealthy_threshold = 10
    enabled             = true
    path                = "/health"
  }
}

resource "aws_lambda_permission" "alb" {
  statement_id  = "AllowExecutionFromALB"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.grpc.function_name
  principal     = "elasticloadbalancing.amazonaws.com"
  source_arn    = aws_lb_target_group.main.arn
}

resource "aws_lb_target_group_attachment" "main" {
  target_group_arn = aws_lb_target_group.main.arn
  target_id        = aws_lambda_function.grpc.arn
  depends_on       = [aws_lambda_permission.alb, aws_lambda_function.grpc]
}
