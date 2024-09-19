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
