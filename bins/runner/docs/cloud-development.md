# Cloud Development

How to build, deploy, and debug the runner management binary on remote cloud runners using AWS SSM.

This is useful when you need to test runner changes on a live cloud instance without going through a full release cycle.

## Cross-Compiling the Binary

The cloud runners run on Amazon Linux 2023 (AL2023) on x86_64 EC2 instances. Build a compatible binary from macOS:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/runner_linux_amd64 ./bins/runner/
```

## Enabling SSM on the Remote Runner

If the EC2 instance isn't connected to SSM, attach the `AmazonSSMManagedInstanceCore` managed policy to its IAM role.

**Find the instance's IAM role:**

```bash
# Get the instance profile ARN
aws ec2 describe-instances \
  --profile <profile> --region <region> \
  --instance-ids <instance-id> \
  --query 'Reservations[*].Instances[*].IamInstanceProfile.Arn'

# Get the role name from the instance profile
aws iam get-instance-profile \
  --profile <profile> \
  --instance-profile-name <profile-name> \
  --query 'InstanceProfile.Roles[*].RoleName'
```

**Attach the SSM policy and reboot:**

```bash
aws iam attach-role-policy \
  --profile <profile> \
  --role-name <role-name> \
  --policy-arn arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore

# Reboot so the SSM agent picks up the new credentials
aws ec2 reboot-instances --profile <profile> --region <region> --instance-ids <instance-id>
```

## Transferring the Binary via S3

Direct SCP over SSM can be complex, so the simplest approach is to use a temporary S3 bucket in the same account as a bridge.

**Upload the binary:**

```bash
# Create a temporary bucket (one-time)
aws s3 mb s3://nuon-dev-tmp --profile <profile> --region <region>

# Upload the binary
aws s3 cp /tmp/runner_linux_amd64 s3://nuon-dev-tmp/runner --profile <profile>
```

**Grant the instance access to the bucket:**

```bash
aws iam put-role-policy \
  --profile <profile> \
  --role-name <role-name> \
  --policy-name tmp-s3-access \
  --policy-document '{
    "Version": "2012-10-17",
    "Statement": [{
      "Effect": "Allow",
      "Action": ["s3:GetObject"],
      "Resource": "arn:aws:s3:::nuon-dev-tmp/*"
    }]
  }'
```

## Replacing the Binary via SSM

The runner management binary lives at `/opt/nuon/runner/bin/runner` and runs as the `nuon-runner-mng` systemd service.

Use `ssm send-command` to stop the service, swap the binary, and restart:

```bash
aws ssm send-command \
  --profile <profile> \
  --region <region> \
  --instance-ids <instance-id> \
  --document-name AWS-RunShellScript \
  --parameters 'commands=[
    "systemctl stop nuon-runner-mng.service",
    "aws s3 cp s3://nuon-dev-tmp/runner /opt/nuon/runner/bin/runner --region <region>",
    "chmod +x /opt/nuon/runner/bin/runner",
    "systemctl start nuon-runner-mng.service",
    "sleep 5",
    "systemctl status nuon-runner-mng.service --no-pager"
  ]'
```

## Verification

**Confirm the binary was replaced by comparing hashes:**

```bash
# Local hash
md5 -q /tmp/runner_linux_amd64

# Remote hash
aws ssm send-command \
  --profile <profile> --region <region> \
  --instance-ids <instance-id> \
  --document-name AWS-RunShellScript \
  --parameters 'commands=["md5sum /opt/nuon/runner/bin/runner"]'
```

**Check service logs:**

```bash
# Systemd journal
aws ssm send-command \
  --profile <profile> --region <region> \
  --instance-ids <instance-id> \
  --document-name AWS-RunShellScript \
  --parameters 'commands=["journalctl -u nuon-runner-mng.service --no-pager -n 50"]'
```

The runner log file is also available at `/var/log/nuon-runner-mng/logs.log`.

## Cleanup

After you're done testing, remove the temporary S3 bucket and inline IAM policy:

```bash
aws s3 rb s3://nuon-dev-tmp --force --profile <profile>
aws iam delete-role-policy --profile <profile> --role-name <role-name> --policy-name tmp-s3-access
```
