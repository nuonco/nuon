<center>
  <img src="https://app.nuon.co/_next/image?url=https%3A%2F%2Favatars.githubusercontent.com%2Fu%2F33817679&w=96&q=75"/>
  <h1>Actions Testing</h1>
  <small>
    AWS | {{ dig "account" "id" "000000000000" .nuon.sandbox.outputs  }} | {{ dig "account" "region" "xx-vvvv-00" .nuon.sandbox.outputs }} | {{ dig "vpc" "id" "vpc-000000" .nuon.sandbox.outputs }}
  </small>

</center>

{{ if .nuon.install_stack.populated }}

## Try it!

Current Cloudformation Status:
<span style="display: inline-block; width: 12px; height: 12px; border-radius: 50%; background-color: {{ if eq
.nuon.install_stack.status "active" }}#2ecc71{{ else if eq .nuon.install_stack.status "error" }}#e74c3c{{ else
}}#f1c40f{{ end }}; margin-right: 5px;"></span>
{{.nuon.install_stack.status}}

- [AWS CloudFormation QuickLink URL]({{.nuon.install_stack.quick_link_url}})
- [AWS CloudFormation Template URL]({{.nuon.install_stack.template_url }})

### Sandbox Mode

You can use sandbox mode end to end, including creating plans. If you have a config that relies on terraform outputs or
something, please up `bins/runner/internal/pkg/jobloop/job_step_outputs.go` to include your outputs so they show up in
sandbox mode and can be rendered.

## README helpers

The README feature includes the set of
[Sprig helper functions](https://masterminds.github.io/sprig/).

{{ "BYOC Retool! " | upper | repeat 5 }}

## Full State

<details>
<summary>👀</summary>
<pre>
{{ toPrettyJson .nuon }}
</pre>
</details>

## Instructions to Access the EKS Cluster

1. Add an access entry for the relevant role.
2. Grant the following perms: AWSEKSAdmin, AWSClusterAdmin
3. Add the cluter kubeconfig w/ the following command.

```bash
aws --region us-west-1 --profile demo.NuonAdmin eks update-kubeconfig --name {{ .nuon.install.id }} --alias {{ .nuon.install.id }}
```
