# Vantage

Terraform to manage account access, users, and CURs.

|      |                                                                        |
| ---- | ---------------------------------------------------------------------- |
| docs | https://docs.vantage.sh/terraform/                                     |
| tf   | https://registry.terraform.io/modules/nuonco/install-access/aws/latest |

## The integration

Involves creating a stack set and and instance of said stack set. We must use
this method otherwise we won't have data from member accounts. This is done in
tf.

## Historical: Out of Band

Vantage was set up ahead of an intro call folloing their docs. The following
commands where executed directly from the terminal.

```bash
aws cloudformation create-stack-set \
  --profile powertoolsdev.NuonAdmin \
  --auto-deployment Enabled=true,RetainStacksOnAccountRemoval=true \
  --permission-mode SERVICE_MANAGED \
  --stack-set-name ConnectToVantage15519-1726084074 \
  --template-url https://vantage-public.s3.amazonaws.com/vantage-integration-nocur-latest.json \
  --parameters \
        ParameterKey=VantageID,ParameterValue=076aa0ef-143a-409c-801c-3f2dc04fda03 \
        ParameterKey=VantageDomain,ParameterValue=https://console.vantage.sh \
        ParameterKey=VantageHandshakeID,ParameterValue=9kY-FCLksNTLxGprJz2rZA \
        ParameterKey=VantagePingbackArn,ParameterValue=arn:aws:sns:us-east-1:630399649041:cross-account-cloudformation-connector \
    ParameterKey=VantageIamRole,ParameterValue=AROAZFRV7IUIYSTS4G3VK \
  --capabilities CAPABILITY_IAM
```

```bash
aws cloudformation create-stack-instances \
  --profile powertoolsdev.NuonAdmin \
  --stack-set-name ConnectToVantage15519-1726084074 \
  --regions us-east-1 \
  --deployment-targets OrganizationalUnitIds=r-p4e3
```

This was undone with the following:

```bash
aws cloudformation delete-stack-instances \
  --profile powertoolsdev.NuonAdmin \
  --stack-set-name ConnectToVantage15519-1726084074 \
  --regions us-east-1 \
  --deployment-targets OrganizationalUnitIds=r-p4e3 \
  --retain-stacks
```

```bash
aws cloudformation delete-stack-set \
  --profile powertoolsdev.NuonAdmin \
  --stack-set-name ConnectToVantage15519-1726084074
```
