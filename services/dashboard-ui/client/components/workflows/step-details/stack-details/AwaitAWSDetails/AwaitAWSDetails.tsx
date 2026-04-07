import { useState } from 'react'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { createFileDownload } from '@/utils/file-download'
import type { IStackDetails } from '../types'

interface IAwaitAWSDetails extends IStackDetails {
  orgId: string
  installId?: string
  installAwsRegion?: string
  hasTerraformInstaller?: boolean
}

export const AwaitAWSDetails = ({
  stack,
  orgId,
  installId,
  installAwsRegion,
  hasTerraformInstaller,
}: IAwaitAWSDetails) => {
  const [isDownloading, setIsDownloading] = useState(false)

  const version = stack?.versions?.at(0)
  const quickLink = version?.quick_link_url
  const templateUrl = version?.template_url
  const isS3Template = templateUrl?.includes('s3.amazonaws.com') || templateUrl?.includes('.s3.')
  const stackName = quickLink?.match(/stackName=([^&]+)/)?.[1] || `nuon-${installId || 'install'}`
  const region = (version as any)?.region || quickLink?.match(/region=([^&#]+)/)?.[1] || installAwsRegion || 'us-east-1'
  const consoleUrl = `https://console.aws.amazon.com/cloudformation/home?region=${region}#/stacks/events?filteringText=${stackName}&filteringStatus=active&viewNested=true`

  const handleDownloadTerraformConfig = async () => {
    if (!orgId || !installId) return

    setIsDownloading(true)
    try {
      const response = await fetch(
        `/api/orgs/${orgId}/installs/${installId}/generate-terraform-installer-config`
      )

      if (!response.ok) {
        throw new Error('Failed to generate terraform installer config')
      }

      const configData = await response.arrayBuffer()
      createFileDownload(
        configData,
        'terraform.tfvars',
        'application/octet-stream'
      )
    } catch (error) {
      console.error('Error downloading terraform installer config:', error)
    } finally {
      setIsDownloading(false)
    }
  }

  return (
    <>
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          {quickLink ? 'Setup your install stack' : 'View your install stack'}
        </Text>

        {quickLink ? (
          <Card>
            <span className="flex justify-between items-center">
              <Text>Quick launch in AWS console</Text>
              <ClickToCopyButton textToCopy={quickLink} />
            </span>
            <Link href={quickLink} target="_blank" rel="noopener noreferrer">
              <Code>{quickLink}</Code>
            </Link>
          </Card>
        ) : null}

        {templateUrl ? (
          <Card>
            <span className="flex justify-between items-center">
              <Text weight="strong">CloudFormation template</Text>
              <span className="flex gap-2 items-center">
                <ClickToCopyButton textToCopy={templateUrl} />
                <Button
                  size="sm"
                  variant="secondary"
                  onClick={() => window.open(templateUrl, '_blank')}
                >
                  Download
                </Button>
              </span>
            </span>
            <Link href={templateUrl} target="_blank" rel="noopener noreferrer">
              <Code>{templateUrl}</Code>
            </Link>
          </Card>
        ) : null}

        {hasTerraformInstaller && (
          <Card>
            <span className="flex justify-between items-center">
              <Text weight="strong">Terraform installer config</Text>
              <Button
                size="sm"
                variant="secondary"
                onClick={handleDownloadTerraformConfig}
                disabled={isDownloading}
              >
                {isDownloading ? 'Downloading...' : 'Download terraform.tfvars'}
              </Button>
            </span>
            <Text variant="subtext">
              Download a pre-configured terraform.tfvars file with your
              install-specific values
            </Text>
          </Card>
        )}
      </div>

      {quickLink ? <Divider dividerWord="or" /> : null}

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Deploy with AWS CLI
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Create stack</Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={isS3Template
                ? `aws cloudformation create-stack --stack-name ${stackName} --template-url ${templateUrl} --capabilities CAPABILITY_NAMED_IAM --region ${region}`
                : `curl -sLo template.json "${templateUrl}" && aws cloudformation create-stack --stack-name ${stackName} --template-body file://template.json --capabilities CAPABILITY_NAMED_IAM --region ${region}`
              }
            />
          </span>
          <Code className="text-xs whitespace-pre-wrap break-all">
            {isS3Template
              ? `aws cloudformation create-stack \\\n  --stack-name ${stackName} \\\n  --template-url ${templateUrl} \\\n  --capabilities CAPABILITY_NAMED_IAM \\\n  --region ${region}`
              : `curl -sLo template.json "${templateUrl}" \\\n  && aws cloudformation create-stack \\\n  --stack-name ${stackName} \\\n  --template-body file://template.json \\\n  --capabilities CAPABILITY_NAMED_IAM \\\n  --region ${region}`
            }
          </Code>
        </Card>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Update existing stack</Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={isS3Template
                ? `aws cloudformation update-stack --stack-name ${stackName} --template-url ${templateUrl} --capabilities CAPABILITY_NAMED_IAM --region ${region}`
                : `curl -sLo template.json "${templateUrl}" && aws cloudformation update-stack --stack-name ${stackName} --template-body file://template.json --capabilities CAPABILITY_NAMED_IAM --region ${region}`
              }
            />
          </span>
          <Code className="text-xs whitespace-pre-wrap break-all">
            {isS3Template
              ? `aws cloudformation update-stack \\\n  --stack-name ${stackName} \\\n  --template-url ${templateUrl} \\\n  --capabilities CAPABILITY_NAMED_IAM \\\n  --region ${region}`
              : `curl -sLo template.json "${templateUrl}" \\\n  && aws cloudformation update-stack \\\n  --stack-name ${stackName} \\\n  --template-body file://template.json \\\n  --capabilities CAPABILITY_NAMED_IAM \\\n  --region ${region}`
            }
          </Code>
        </Card>
      </div>

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Verify your stack
        </Text>
        <Card>
          <Text variant="subtext" theme="neutral">
            After running the create or update command above, open the AWS CloudFormation console to monitor your stack progress.
          </Text>
          <Button
            variant="secondary"
            onClick={() => window.open(consoleUrl, '_blank')}
          >
            Open in AWS console
          </Button>
        </Card>
      </div>
    </>
  )
}

export const AwaitAWSDetailsSkeleton = () => {
  return (
    <>
      <Skeleton height="24px" width="175px" />

      <Card>
        <Skeleton height="17px" width="100px" />
        <Skeleton height="132px" width="100%" />
      </Card>

      <Card>
        <Skeleton height="17px" width="120px" />
        <Skeleton height="72px" width="100%" />
      </Card>

      <Card>
        <Skeleton height="17px" width="175px" />
        <Skeleton height="32px" width="100%" />
      </Card>

      <Skeleton height="24px" width="325px" />

      <Card>
        <Skeleton height="17px" width="219px" />
        <Skeleton height="92px" width="100%" />
      </Card>

      <Skeleton height="24px" width="382px" />

      <Card>
        <Skeleton height="17px" width="223px" />
        <Skeleton height="92px" width="100%" />
      </Card>
    </>
  )
}
