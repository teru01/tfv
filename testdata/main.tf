resource "aws_ecs_cluster" "foovar" {
  name = "hello"
}

resource "aws_ecs_service" "mongo" {
  name            = "mongodb-${var.env}"
  cluster         = aws_ecs_cluster.foovar.id
  task_definition = "taskdef_arn"
  desired_count   = 3
  iam_role        = var.iam_role_arn

  load_balancer {
    target_group_arn = "tg_arn"
    container_name   = "mongo-${var.appname}"
    container_port   = 8080
  }

  placement_constraints {
    type       = "memberOf"
    expression = "attribute:ecs.availability-zone in [us-west-2a, us-west-2b]"
  }
}

locals {
  foo = "Hello, %{if var.name != "%{if var.name != "${var.moge}"}${var.name}%{else}unnamed%{endif}"}${var.name}%{else}unnamed%{endif}!"
}
