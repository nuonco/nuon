'use client'

import { useParams } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { ClickToCopyButton } from '@/components/ClickToCopy'
import { ConfigurationVariables } from '@/components/ComponentConfig'
import { Link } from '@/components/Link'
import { Loading } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { StatusBadge } from '@/components/Status'
import { Time } from '@/components/Time'
import { Text, Code } from '@/components/Typography'
import type { TAppStackConfig, TInstallStack } from '@/types'
import type { IPollStepDetails } from './InstallWorkflowSteps'

interface IStackStepDetails extends IPollStepDetails {
  appId: string
}

export const StackStep: FC<IStackStepDetails> = ({
  appId,
  step,
  shouldPoll = false,
  pollDuration = 5000,
}) => {
  const isGenerateStep = step?.name === 'generate install stack'
  const params = useParams<Record<'org-id', string>>()
  const orgId = params?.['org-id']
  const [stack, setData] = useState<TInstallStack>()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()

  const fetchData = () => {
    fetch(`/api/${orgId}/installs/${step?.install_id}/stack`).then((r) =>
      r.json().then((res) => {
        setIsLoading(false)
        if (res?.error) {
          setError(res?.error?.error)
        } else {
          setData(res.data)
        }
      })
    )
  }

  useEffect(() => {
    fetchData()
  }, [])

  useEffect(() => {
    if (shouldPoll) {
      const pollData = setInterval(fetchData, pollDuration)

      return () => clearInterval(pollData)
    }
  }, [shouldPoll])

  return (
    <>
      {isLoading ? (
        <Loading loadingText="Loading deploy details..." variant="page" />
      ) : (
        <>
          {error ? <Notice>{error}</Notice> : null}
          {stack ? (
            isGenerateStep ? (
              <GenerateStack stack={stack} appId={appId} orgId={orgId} />
            ) : (
              <AwaitStack stack={stack} />
            )
          ) : null}
        </>
      )}
    </>
  )
}

const GenerateStack: FC<{
  stack: TInstallStack
  appId: string
  orgId: string
}> = ({ stack, appId, orgId }) => {
  const [stackConfig, setStackConfig] = useState<TAppStackConfig>()
  const [isLoading, setIsLoading] = useState(!Boolean(stackConfig))
  const [error, setError] = useState<string>()

  const fetchStackConfig = () => {
    fetch(
      `/api/${orgId}/apps/${appId}/configs/${stack?.versions?.at(-1).app_config_id}`
    ).then((r) =>
      r.json().then((res) => {
        setIsLoading(false)
        if (res?.error) {
          setError(res?.error?.error)
        } else {
          setStackConfig(res.data?.stack)
        }
      })
    )
  }

  useEffect(() => {
    if (stack?.versions && !stackConfig) {
      fetchStackConfig()
    }
  }, [stack])

  return (
    <>
      {isLoading ? (
        <Loading loadingText="Loading stack infromation..." variant="page" />
      ) : (
        <>
          {error ? <Notice>{error}</Notice> : null}
          {stack ? (
            <>
              <>
                <div className="flex flex-col gap-2">
                  <Text variant="med-14">Setup your install stack</Text>
                  <div className="border rounded-lg shadow-sm p-2 flex flex-col gap-1">
                    <span className="flex justify-between items-center">
                      <Text variant="med-12">Install quick link</Text>
                      <ClickToCopyButton
                        textToCopy={stack?.versions?.at(-1)?.quick_link_url}
                      />
                    </span>
                    <Code>{stack?.versions?.at(-1)?.quick_link_url}</Code>
                  </div>
                </div>

                <div className="relative">
                  <hr />
                  <Text className="shadow-sm px-2 border w-fit rounded-lg bg-white text-cool-grey-950 dark:bg-dark-grey-100 dark:text-cool-grey-50 absolute inset-0 m-auto h-[18px]">
                    OR
                  </Text>
                </div>

                <div className="flex flex-col gap-2">
                  <Text variant="med-14">
                    Setup your install stack using CLI command
                  </Text>
                  <div className="border rounded-lg shadow-sm p-2 flex flex-col gap-1">
                    <ClickToCopyButton
                      className="w-fit self-end"
                      textToCopy={stack?.versions?.at(-1)?.quick_link_url}
                    />
                    <Code>
                      aws cloudformation create-stack --stack-name
                      [YOUR_STACK_NAME] --template-url{' '}
                      {stack?.versions?.at(-1)?.template_url}
                    </Code>
                  </div>
                </div>

                <div className="flex flex-col gap-2">
                  <Text variant="med-14">
                    Update an existing install stack using CLI command
                  </Text>
                  <div className="border rounded-lg shadow-sm p-2 flex flex-col gap-1">
                    <ClickToCopyButton
                      className="w-fit self-end"
                      textToCopy={stack?.versions?.at(-1)?.quick_link_url}
                    />
                    <Code>
                      aws cloudformation update-stack --stack-name
                      [YOUR_STACK_NAME] --template-url{' '}
                      {stack?.versions?.at(-1)?.template_url}
                    </Code>
                  </div>
                </div>

                {stackConfig ? (
                  <div className="flex flex-col gap-2">
                    <ConfigurationVariables
                      heading="Stack template details"
                      headingVariant="med-14"
                      variables={{
                        name: stackConfig?.name,
                        description: stackConfig?.description,
                        runner_nested_template_url:
                          stackConfig?.runner_nested_template_url,
                        vpc_nested_template_url:
                          stackConfig?.vpc_nested_template_url,
                        type: stackConfig?.type,
                      }}
                      isNotTruncated
                    />
                  </div>
                ) : null}
              </>
            </>
          ) : null}
        </>
      )}
    </>
  )
}

const AwaitStack: FC<{ stack: TInstallStack }> = ({ stack }) => {
  return (
    <>
      <div className="border rounded-lg shadow-sm flex flex-col gap-1">
        <div className="flex justify-between items-center px-2 py-4 border-b">
          <Text variant="med-12">
            Install stack{' '}
            {stack?.versions?.at(-1)?.composite_status?.status === 'active'
              ? 'up and running'
              : 'is waiting to run'}
          </Text>
        </div>
        <div className="px-2 py-4 grid grid-cols-4">
          <StatusBadge
            status={stack?.versions?.at(-1)?.composite_status?.status}
            description={
              stack?.versions?.at(-1)?.composite_status
                ?.status_human_description
            }
            label="Current status"
          />
          <div className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Last checked
            </Text>
            <Time
              time={stack?.versions?.at(-1).runs?.at(-1)?.updated_at}
              format="relative"
            />
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-2">
        <ConfigurationVariables
          heading="Stack outputs"
          headingVariant="med-14"
          variables={stack?.install_stack_outputs?.data}
          isNotTruncated
        />
      </div>
    </>
  )
}
