package cloudformation

import (
	"fmt"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/iam"
	"github.com/awslabs/goformation/v7/cloudformation/logs"
	"github.com/awslabs/goformation/v7/cloudformation/ssm"
)

func (a *Templates) getRunnerCloudWatchLogGroup(inp *TemplateInput, t tagBuilder) *logs.LogGroup {
	return &logs.LogGroup{
		LogGroupName:    ptr(fmt.Sprintf("runner-%s", inp.Runner.ID)),
		RetentionInDays: ptr(7),
		Tags:            t.apply(nil, "runner-cw-lg"),
	}
}

func (a *Templates) getRunnerCloudWatchLogStream(inp *TemplateInput, t tagBuilder) *logs.LogStream {
	// create a default cloudwatch logs stream
	return &logs.LogStream{
		LogGroupName:  cloudformation.Ref("RunnerCloudWatchLogGroup"),
		LogStreamName: ptr(fmt.Sprintf("runner-%s", inp.Runner.ID)),
	}
}

func (a *Templates) getRunnerCloudWatchLogPolicy(inp *TemplateInput, t tagBuilder) *iam.Policy {
	// a policy foro the runner instance role so it can write logs to the log group defined below
	// src: https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/iam-identity-based-access-control-cwl.html#w292aac43c15c15c25c13

	return &iam.Policy{
		PolicyName: fmt.Sprintf("nuon-install-%s-cw-logs-access", inp.Install.ID),
		Roles: []string{
			cloudformation.GetAtt("RunnerAutoScalingGroup", "Outputs.RunnerInstanceRole"),
		},
		PolicyDocument: map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []interface{}{
				map[string]interface{}{
					"Action": []string{
						"logs:CreateLogStream",
						"logs:PutLogEvents",
					},
					"Effect": "Allow",
					"Resource": []interface{}{
						cloudformation.GetAtt("RunnerCloudWatchLogGroup", "Arn"),
					},
				},
			},
		},
	}
}

func (a *Templates) getRunnerCloudWatchAgentConfig(inp *TemplateInput, t tagBuilder) *ssm.Parameter {
	// NOTE: idk if we're actually using this rn - configure the log group to send logs from `journalctl -f -u nuon-runner` to cloudwatch
	return &ssm.Parameter{
		Name:        ptr(fmt.Sprintf("runner-cw-cfg-%s", inp.Runner.ID)),
		Type:        "String",
		Description: ptr("CloudWatch agent configuration for both metrics and logs"),
		Value: `{
          "agent": {
            "metrics_collection_interval": 60,
            "run_as_user": "root"
          },
          "metrics": {
            "metrics_collected": {
              "disk": {
                "measurement": ["used_percent"],
                "resources": ["/"],
                "drop_device": true
              },
              "mem": {
                "measurement": ["mem_used_percent"]
              },
              "swap": {
                "measurement": ["swap_used_percent"]
              }
            },
            "append_dimensions": {
              "AutoScalingGroupName": "${aws:AutoScalingGroupName}"
            }
          },
          "logs": {
            "logs_collected": {
              "files": {
                "collect_list": [
                  {
                    "file_path": "/var/log/messages",
                    "log_group_name": "/ec2/messages",
                    "log_stream_name": "{instance_id}"
                  },
                  {
                    "file_path": "/var/log/secure",
                    "log_group_name": "/ec2/secure",
                    "log_stream_name": "{instance_id}"
                  },
                  {
                    "file_path": "/var/log/httpd/access_log",
                    "log_group_name": "/ec2/httpd/access_log",
                    "log_stream_name": "{instance_id}"
                  },
                  {
                    "file_path": "/var/log/httpd/error_log",
                    "log_group_name": "/ec2/httpd/error_log",
                    "log_stream_name": "{instance_id}"
                  }
                ]
              }
            }
          }
        }`,
	}
}
