import { useMemo, useState } from 'react'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { createFileDownload } from '@/utils/file-download'
import type { IStackDetails } from '../types'

interface IAwaitAWSDetails extends IStackDetails {
  orgId: string
  installId?: string
  installAwsRegion?: string
}

// The tfvars envelope ctl-api stores in `terraform_contents` is a JSON
// document of shape `{"tfvars": "<hcl body>"}`. Mirrors the GCP parser at
// AwaitGCPDetails.tsx.
function parseTfvars(contents: unknown): string {
  if (!contents) return ''

  let raw: unknown = contents
  if (typeof raw === 'string') {
    try {
      raw = JSON.parse(raw)
    } catch {
      try {
        raw = JSON.parse(atob(raw as string))
      } catch {
        return ''
      }
    }
  }

  if (typeof raw === 'object' && raw !== null && 'tfvars' in raw) {
    return String((raw as Record<string, unknown>).tfvars ?? '')
  }

  return ''
}

export const AwaitAWSDetails = ({
  stack,
  orgId,
  installId,
  installAwsRegion,
}: IAwaitAWSDetails) => {
  const version = stack?.versions?.at(0)
  // The new TerraformContents fields aren't in the regenerated OpenAPI types
  // yet; bridge with a local widening cast.
  const versionExt = version as
    | (typeof version & {
        terraform_contents?: unknown
        terraform_checksum?: string
      })
    | undefined

  const tfvarsContent = useMemo(
    () => parseTfvars(versionExt?.terraform_contents),
    [versionExt?.terraform_contents]
  )
  const hasTerraform = tfvarsContent.length > 0

  return (
    <div className="flex flex-col gap-4">
      <Text variant="base" weight="strong">
        Setup your install stack
      </Text>

      {hasTerraform ? (
        <Tabs
          initActiveTab="cloudformation"
          tabs={{
            cloudformation: (
              <CloudFormationTab
                version={version}
                installId={installId}
                installAwsRegion={installAwsRegion}
              />
            ),
            terraform: (
              <TerraformTab
                tfvarsContent={tfvarsContent}
                orgId={orgId}
                installId={installId}
              />
            ),
          }}
        />
      ) : (
        <CloudFormationTab
          version={version}
          installId={installId}
          installAwsRegion={installAwsRegion}
        />
      )}
    </div>
  )
}

interface ICloudFormationTab {
  version: NonNullable<IStackDetails['stack']['versions']>[number] | undefined
  installId?: string
  installAwsRegion?: string
}

const CloudFormationTab = ({
  version,
  installId,
  installAwsRegion,
}: ICloudFormationTab) => {
  const quickLink = version?.quick_link_url
  const templateUrl = version?.template_url
  const isS3Template =
    templateUrl?.includes('s3.amazonaws.com') || templateUrl?.includes('.s3.')
  const stackName =
    quickLink?.match(/stackName=([^&]+)/)?.[1] ||
    `nuon-${installId || 'install'}`
  const region =
    (version as { region?: string } | undefined)?.region ||
    quickLink?.match(/region=([^&#]+)/)?.[1] ||
    installAwsRegion ||
    ''
  // CLI commands and console links work whether the user already chose a
  // region or hasn't yet — when unknown, render a `<YOUR_REGION>` placeholder
  // and let the user substitute at run-time.
  const regionForCmd = region || '<YOUR_REGION>'
  const consoleUrl = region
    ? `https://console.aws.amazon.com/cloudformation/home?region=${region}#/stacks/events?filteringText=${stackName}&filteringStatus=active&viewNested=true`
    : `https://console.aws.amazon.com/cloudformation/home#/stacks?filteringText=${stackName}`

  return (
    <div className="flex flex-col gap-4 pt-4">
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
      ) : (
        <Card>
          <Text variant="subtext" theme="neutral">
            Open the AWS console in your preferred region, then create a
            CloudFormation stack from the template URL below. The CLI snippets
            further down include a <code>&lt;YOUR_REGION&gt;</code> placeholder
            you can substitute.
          </Text>
        </Card>
      )}

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
              textToCopy={
                isS3Template
                  ? `aws cloudformation create-stack --stack-name ${stackName} --template-url ${templateUrl} --capabilities CAPABILITY_NAMED_IAM --region ${regionForCmd}`
                  : `curl -sLo template.json "${templateUrl}" && aws cloudformation create-stack --stack-name ${stackName} --template-body file://template.json --capabilities CAPABILITY_NAMED_IAM --region ${regionForCmd}`
              }
            />
          </span>
          <Code className="text-xs whitespace-pre-wrap break-all">
            {isS3Template
              ? `aws cloudformation create-stack \\\n  --stack-name ${stackName} \\\n  --template-url ${templateUrl} \\\n  --capabilities CAPABILITY_NAMED_IAM \\\n  --region ${regionForCmd}`
              : `curl -sLo template.json "${templateUrl}" \\\n  && aws cloudformation create-stack \\\n  --stack-name ${stackName} \\\n  --template-body file://template.json \\\n  --capabilities CAPABILITY_NAMED_IAM \\\n  --region ${regionForCmd}`}
          </Code>
        </Card>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Update existing stack</Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={
                isS3Template
                  ? `aws cloudformation update-stack --stack-name ${stackName} --template-url ${templateUrl} --capabilities CAPABILITY_NAMED_IAM --region ${regionForCmd}`
                  : `curl -sLo template.json "${templateUrl}" && aws cloudformation update-stack --stack-name ${stackName} --template-body file://template.json --capabilities CAPABILITY_NAMED_IAM --region ${regionForCmd}`
              }
            />
          </span>
          <Code className="text-xs whitespace-pre-wrap break-all">
            {isS3Template
              ? `aws cloudformation update-stack \\\n  --stack-name ${stackName} \\\n  --template-url ${templateUrl} \\\n  --capabilities CAPABILITY_NAMED_IAM \\\n  --region ${regionForCmd}`
              : `curl -sLo template.json "${templateUrl}" \\\n  && aws cloudformation update-stack \\\n  --stack-name ${stackName} \\\n  --template-body file://template.json \\\n  --capabilities CAPABILITY_NAMED_IAM \\\n  --region ${regionForCmd}`}
          </Code>
        </Card>
      </div>

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Verify your stack
        </Text>
        <Card>
          <Text variant="subtext" theme="neutral">
            After running the create or update command above, open the AWS
            CloudFormation console to monitor your stack progress.
          </Text>
          <Button
            variant="secondary"
            onClick={() => window.open(consoleUrl, '_blank')}
          >
            Open in AWS console
          </Button>
        </Card>
      </div>
    </div>
  )
}

interface ITerraformTab {
  tfvarsContent: string
  orgId: string
  installId?: string
}

const TerraformTab = ({
  tfvarsContent,
  orgId,
  installId,
}: ITerraformTab) => {
  const [isDownloading, setIsDownloading] = useState(false)

  const handleDownload = async () => {
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
        'install.tfvars',
        'application/octet-stream'
      )
    } catch (error) {
      console.error('Error downloading terraform installer config:', error)
    } finally {
      setIsDownloading(false)
    }
  }

  const cloneCmd = `git clone https://github.com/nuonco/install-stacks.git
cd install-stacks/aws`

  const backendSnippet = `terraform {
  backend "s3" {
    bucket = "<your-state-bucket>"
    key    = "nuon/${installId}/terraform.tfstate"
    region = "<your-state-bucket-region>"
  }
}`

  const applyCmd = `terraform init && terraform apply -var-file=install.tfvars`

  return (
    <div className="flex flex-col gap-4 pt-4">

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          1. Clone the install stack module
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Clone and enter the AWS module directory</Text>
            <ClickToCopyButton textToCopy={cloneCmd} />
          </span>
          <Code variant="preformated">{cloneCmd}</Code>
        </Card>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          2. Configure remote state (recommended)
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Create a <code>backend.tf</code> file to store Terraform state in
              S3
            </Text>
            <ClickToCopyButton textToCopy={backendSnippet} />
          </span>
          <Code variant="preformated">{backendSnippet}</Code>
        </Card>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          3. Save the install configuration
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Save this as <code>install.tfvars</code>, or download it
              directly
            </Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={tfvarsContent} />
              <Button
                size="sm"
                variant="secondary"
                onClick={handleDownload}
                disabled={isDownloading}
              >
                {isDownloading ? 'Downloading...' : 'Download'}
              </Button>
            </span>
          </span>
          <Code variant="preformated">{tfvarsContent}</Code>
        </Card>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          4. Apply with Terraform
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Run Terraform</Text>
            <ClickToCopyButton textToCopy={applyCmd} />
          </span>
          <Code variant="preformated">{applyCmd}</Code>
        </Card>
      </div>
    </div>
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
