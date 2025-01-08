# aws cloudformation create-stack-set \
#   --profile powertoolsdev.NuonAdmin \
#   --auto-deployment Enabled=true,RetainStacksOnAccountRemoval=true \
#   --permission-mode SERVICE_MANAGED \
#   --stack-set-name ConnectToVantage15519-1726084074 \
#   --template-url https://vantage-public.s3.amazonaws.com/vantage-integration-nocur-latest.json \
#   --parameters \
#         ParameterKey=VantageID,ParameterValue=076aa0ef-143a-409c-801c-3f2dc04fda03 \
#         ParameterKey=VantageDomain,ParameterValue=https://console.vantage.sh \
#         ParameterKey=VantageHandshakeID,ParameterValue=9kY-FCLksNTLxGprJz2rZA \
#         ParameterKey=VantagePingbackArn,ParameterValue=arn:aws:sns:us-east-1:630399649041:cross-account-cloudformation-connector \
#     ParameterKey=VantageIamRole,ParameterValue=AROAZFRV7IUIYSTS4G3VK \
#   --capabilities CAPABILITY_IAM

resource "aws_cloudformation_stack_set" "vantage-integrations-stack-set" {
  name         = "ConnectToVantage15519-1726084074"
  template_url = "https://vantage-public.s3.amazonaws.com/vantage-integration-nocur-latest.json"

  permission_model = "SERVICE_MANAGED"
  auto_deployment {
    enabled                          = true
    retain_stacks_on_account_removal = true
  }

  parameters = {
    VantageID          = "076aa0ef-143a-409c-801c-3f2dc04fda03"
    VantageDomain      = "https://console.vantage.sh"
    VantageHandshakeID = "9kY-FCLksNTLxGprJz2rZA"
    VantagePingbackArn = "arn:aws:sns:us-east-1:630399649041:cross-account-cloudformation-connector"
    VantageIamRole     = "AROAZFRV7IUIYSTS4G3VK"
  }
  capabilities = ["CAPABILITY_IAM"]
}

resource "aws_cloudformation_stack" "connect-to-vantage" {
  name         = "ConnectToVantage15519-1726781115"
  template_url = "https://vantage-public.s3.amazonaws.com/vantage-integration-combined-latest.json"

  parameters = {
    BucketName                  = "vantage-cur-076aa0ef-143a-409c-801c-3f2dc04fda03-226346b98e"
    ReportName                  = "VantageReport-226346b98e"
    VantageDomain               = "https://console.vantage.sh"
    VantageHandshakeID          = "9kY-FCLksNTLxGprJz2rZA"
    VantageID                   = "076aa0ef-143a-409c-801c-3f2dc04fda03"
    VantageIamRole              = "AROAZFRV7IUIYSTS4G3VK"
    VantageNotificationTopicArn = "arn:aws:sns:us-east-1:630399649041:cost-and-usage-report-uploaded"
    VantagePingbackArn          = "arn:aws:sns:us-east-1:630399649041:cross-account-cloudformation-connector"
  }

  capabilities = ["CAPABILITY_IAM"]
}

# aws cloudformation create-stack-instances \
#   --profile powertoolsdev.NuonAdmin \
#   --stack-set-name ConnectToVantage15519-1726084074 \
#   --regions us-east-1 \
#   --deployment-targets OrganizationalUnitIds=r-p4e3

resource "aws_cloudformation_stack_set_instance" "vantage-integrations-stack-set-instance" {
  region         = "us-east-1"
  stack_set_name = aws_cloudformation_stack_set.vantage-integrations-stack-set.name

  deployment_targets {
    organizational_unit_ids = ["r-p4e3"]
  }
}

#
# upgrade vantage: Wed Jan 8, 2025
#
# NOTE(fd): do we have to do anything to remove the old cloudformation stacks? i think we should delete them after this is done.
# aws cloudformation create-stack --stack-name ConnectToVantage15519-1736350314 \
#   --template-url https://vantage-public.s3.amazonaws.com/vantage-integration-update-latest.json \
#   --parameters ParameterKey=VantageCrossAccountRole,ParameterValue='ConnectToVantage15519-1726781115-CrossAccountRole-13HVDF0Jjh9K' \
#   --capabilities CAPABILITY_IAM \
#   --region us-east-1
resource "aws_cloudformation_stack" "update-vantage" {
  name         = "ConnectToVantage15519-1736347644"
  template_url = "https://vantage-public.s3.amazonaws.com/vantage-integration-update-latest.json"

  parameters = {
    VantageCrossAccountRole = "ConnectToVantage15519-1726781115-CrossAccountRole-13HVDF0Jjh9K"
  }

  capabilities = ["CAPABILITY_IAM"]
}

# per resource costs
# aws cloudformation create-stack --stack-name VantageSetupCUR30834-1736350314 \
#   --template-url https://vantage-public.s3.amazonaws.com/cur-setup-latest.json \
#   --parameters ParameterKey=VantageID,ParameterValue='O1XMAv_xX3ZE_wvwyHtrsg' \
#     ParameterKey=VantageDomain,ParameterValue='https://console.vantage.sh' \
#     ParameterKey=VantagePingbackArn,ParameterValue='arn:aws:sns:us-east-1:630399649041:cost-and-usage-report-connector' \
#     ParameterKey=VantageIamRole,ParameterValue='ConnectToVantage15519-1726781115-CrossAccountRole-13HVDF0Jjh9K' \
#     ParameterKey=VantageIamRoleArn,ParameterValue='arn:aws:iam::491187160246:role/ConnectToVantage15519-1726781115-CrossAccountRole-13HVDF0Jjh9K' \
#     ParameterKey=VantageNotificationTopicArn,ParameterValue='arn:aws:sns:us-east-1:630399649041:cost-and-usage-report-uploaded' \
#     ParameterKey=ReportName,ParameterValue='VantageReport-2549feb6a9' \
#     ParameterKey=BucketName,ParameterValue='vantage-cur-o1xmavxx3zewvwyhtrsg-2549feb6a9' \
#   --capabilities CAPABILITY_IAM \
#   --region us-east-1
resource "aws_cloudformation_stack" "per-resource-costs" {
  name         = "VantageSetupCUR30834-1736347644"
  template_url = "https://vantage-public.s3.amazonaws.com/cur-setup-latest.json"

  parameters = {
    VantageCrossAccountRole     = "ConnectToVantage15519-1726781115-CrossAccountRole-13HVDF0Jjh9K"
    VantageID                   = "O1XMAv_xX3ZE_wvwyHtrsg"
    VantageDomain               = "https://console.vantage.sh"
    VantagePingbackArn          = "arn:aws:sns:us-east-1:630399649041:cost-and-usage-report-connector"
    VantageIamRole              = "ConnectToVantage15519-1726781115-CrossAccountRole-13HVDF0Jjh9K"
    VantageIamRoleArn           = "arn:aws:iam::491187160246:role/ConnectToVantage15519-1726781115-CrossAccountRole-13HVDF0Jjh9K"
    VantageNotificationTopicArn = "arn:aws:sns:us-east-1:630399649041:cost-and-usage-report-uploaded"
    ReportName                  = "VantageReport-88e7dfd029"
    BucketName                  = "vantage-cur-o1xmavxx3zewvwyhtrsg-88e7dfd029"
  }

  capabilities = ["CAPABILITY_IAM"]

  depends_on = [aws_cloudformation_stack.update-vantage]
}
