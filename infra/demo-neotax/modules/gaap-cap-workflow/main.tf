resource "aws_sqs_queue" "step_0_coordinator" {
  name = "step_0_coordinator"
}

resource "aws_sqs_queue" "step_0_children" {
  name = "step_0_children"
}

resource "aws_sqs_queue" "step_1_coordinator" {
  name = "step_1_coordinator"
}
