import { Button } from '@/components/common/Button'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Link } from '@/components/old/Link'
import { Text } from '@/components/common/Text'

export const StackLinks = ({
  template_url,
  quick_link_url,
}: {
  template_url: string
  quick_link_url: string
}) => {
  return (
    <>
      {quick_link_url ? (
        <div className="border rounded-md shadow p-2 flex flex-col gap-1">
          <span className="flex justify-between items-center">
            <Text variant="body" weight="strong">
              Install quick link
            </Text>
            <ClickToCopyButton textToCopy={quick_link_url} />
          </span>
          <Link href={quick_link_url} target="_blank" rel="noopener noreferrer">
            <Code>{quick_link_url}</Code>
          </Link>
        </div>
      ) : null}

      {template_url ? (
        <div className="border rounded-md shadow p-2 flex flex-col gap-3 mt-3">
          <span className="flex justify-between items-center">
            <Text variant="body" weight="strong">
              Install template
            </Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={template_url} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() => window.open(template_url, '_blank')}
              >
                Download
              </Button>
            </span>
          </span>
          <Link href={template_url} target="_blank" rel="noopener noreferrer">
            <Code>{template_url}</Code>
          </Link>

          {!quick_link_url ? (
            <div className="border-t pt-3 flex flex-col gap-2">
              <Text variant="subtext" weight="strong">
                Deploy with AWS CLI
              </Text>
              <Text variant="subtext" theme="neutral">
                Download the template and create the stack using the AWS CLI:
              </Text>
              <div className="relative">
                <ClickToCopyButton
                  textToCopy={`curl -sLo template.json "${template_url}" && aws cloudformation create-stack --stack-name nuon-install --template-body file://template.json --capabilities CAPABILITY_NAMED_IAM --region us-east-1`}
                />
                <Code className="text-xs overflow-x-auto block whitespace-pre">
{`curl -sLo template.json "${template_url}" \\
  && aws cloudformation create-stack \\
    --stack-name nuon-install \\
    --template-body file://template.json \\
    --capabilities CAPABILITY_NAMED_IAM \\
    --region us-east-1`}
                </Code>
              </div>
            </div>
          ) : null}
        </div>
      ) : null}
    </>
  )
}
