
output "function_name" {
  description = "Name of the Lambda function."

  value = aws_lambda_function.grpc.function_name
}

output "alb" {
  description = "DNS for the ALB"

  value = aws_lb.alb.dns_name
}
