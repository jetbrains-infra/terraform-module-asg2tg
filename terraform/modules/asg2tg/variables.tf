variable "project_name" {
  type = string
}
variable "stage" {
  type = string
}
variable "name" {
  type = string
}
variable "tags" {
  type    = "map"
  default = {}
}
variable "target_group_arn" {
  type = string
}
variable "target_group_port" {
  type = number
}
variable "autoscaling_groups_names" {
  type = "list"
}

locals {
  dashed_name = "${lower(replace(var.project_name, " ", "-"))}-${lower(replace(var.stage, " ", "-"))}"
  ebs_id      = "${local.dashed_name}-${var.name}"
  tags = merge(var.tags, {
    Name    = local.dashed_name
    Service = var.name
  })
}