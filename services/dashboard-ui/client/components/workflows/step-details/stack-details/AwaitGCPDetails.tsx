'use client'

import { useState, useMemo } from 'react'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { api } from '@/lib/api'
import type { IStackDetails } from './types'

function parseTfvars(contents: unknown): string {
  if (!contents) return ''

  let raw = contents
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

export const AwaitGCPDetails = ({ stack }: IStackDetails) => {
  const { install } = useInstall()
  const { org } = useOrg()
  const [tokenVisible, setTokenVisible] = useState(false)
  const [runnerApiToken, setRunnerApiToken] = useState('')
  const [expiresAt, setExpiresAt] = useState('')
  const [tokenError, setTokenError] = useState('')
  const [isGenerating, setIsGenerating] = useState(false)

  const version = stack?.versions?.at(0)
  const tfvarsContent = useMemo(
    () => parseTfvars(version?.contents),
    [version?.contents]
  )

  const generateToken = async () => {
    if (!org?.id || !install?.id) return

    setIsGenerating(true)
    setTokenError('')
    try {
      const data = await api<{ token: string; expires_at: string }>({
        method: 'POST',
        orgId: org.id,
        path: `installs/${install.id}/runner-bootstrap-token`,
      })
      setRunnerApiToken(data?.token ?? '')
      setExpiresAt(data?.expires_at ?? '')
    } catch (error: any) {
      setTokenError(error?.error || 'Failed to generate token')
    } finally {
      setIsGenerating(false)
    }
  }

  const maskedToken = runnerApiToken
    ? `${runnerApiToken.slice(0, 8)}${'•'.repeat(24)}`
    : ''

  const applyCmd = runnerApiToken
    ? `TF_VAR_runner_api_token="${runnerApiToken}" \\
  terraform init && terraform apply -var-file=install.tfvars`
    : `TF_VAR_runner_api_token="<generate token above>" \\
  terraform init && terraform apply -var-file=install.tfvars`

  const displayApplyCmd = runnerApiToken
    ? `TF_VAR_runner_api_token="${maskedToken}" \\
  terraform init && terraform apply -var-file=install.tfvars`
    : applyCmd

  const cloneCmd = `git clone https://github.com/nuonco/install-stacks.git
cd install-stacks/gcp`

  const backendSnippet = `terraform {
  backend "gcs" {
    bucket = "<your-state-bucket>"
    prefix = "nuon/${install?.id}"
  }
}`

  return (
    <>
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          1. Clone the install stack module
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Clone and enter the GCP module directory</Text>
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
              GCS
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
              Save this as <code>install.tfvars</code>
            </Text>
            <ClickToCopyButton textToCopy={tfvarsContent} />
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
          <Text variant="subtext">
            Generate a runner API token below. Each token expires in 2 hours —
            click again to rotate.
          </Text>

          <span className="flex justify-between items-center">
            <Text>Runner API token</Text>
            <span className="flex items-center gap-1">
              {runnerApiToken && (
                <>
                  <button
                    type="button"
                    className="hover:bg-black/10 dark:hover:bg-white/5 flex items-center cursor-pointer border rounded-md p-1"
                    onClick={() => setTokenVisible((v) => !v)}
                    aria-label={tokenVisible ? 'Hide token' : 'Reveal token'}
                  >
                    <Icon
                      variant={tokenVisible ? 'EyeSlash' : 'Eye'}
                      size="16"
                    />
                  </button>
                  <ClickToCopyButton textToCopy={runnerApiToken} />
                </>
              )}
              <Button
                size="sm"
                variant="secondary"
                onClick={generateToken}
                disabled={isGenerating || !org?.id || !install?.id}
              >
                {isGenerating
                  ? 'Generating...'
                  : runnerApiToken
                    ? 'Rotate token'
                    : 'Generate token'}
              </Button>
            </span>
          </span>

          {tokenError && (
            <Text variant="subtext">{tokenError}</Text>
          )}

          {runnerApiToken && (
            <>
              <Code variant="preformated">
                {tokenVisible ? runnerApiToken : maskedToken}
              </Code>
              {expiresAt && (
                <Text variant="subtext">
                  Expires: {new Date(expiresAt).toLocaleString()}
                </Text>
              )}
            </>
          )}

          <Divider />

          <span className="flex justify-between items-center">
            <Text>Run Terraform</Text>
            <ClickToCopyButton textToCopy={applyCmd} />
          </span>
          <Code variant="preformated">{displayApplyCmd}</Code>
        </Card>
      </div>
    </>
  )
}

export const AwaitGCPDetailsSkeleton = () => {
  return (
    <>
      <Skeleton height="24px" width="275px" />

      <Card>
        <Skeleton height="17px" width="250px" />
        <Skeleton height="52px" width="100%" />
      </Card>

      <Divider />

      <Skeleton height="24px" width="300px" />

      <Card>
        <Skeleton height="17px" width="300px" />
        <Skeleton height="72px" width="100%" />
      </Card>

      <Divider />

      <Skeleton height="24px" width="250px" />

      <Card>
        <Skeleton height="17px" width="200px" />
        <Skeleton height="100px" width="100%" />
      </Card>

      <Divider />

      <Skeleton height="24px" width="200px" />

      <Card>
        <Skeleton height="17px" width="150px" />
        <Skeleton height="52px" width="100%" />
      </Card>
    </>
  )
}
