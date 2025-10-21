'use client'

import { ClickToCopyButton } from '@/components/ClickToCopy'
import { ConfigurationVariables } from '@/components/ComponentConfig'
import { Link } from '@/components/Link'
import { Loading } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { StatusBadge } from '@/components/Status'
import { Time } from '@/components/Time'
import { Text, Code } from '@/components/Typography'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling } from '@/hooks/use-polling'
import { useQuery } from '@/hooks/use-query'
import type { TAppStackConfig, TInstallStack } from '@/types'
import type { IPollStepDetails } from './InstallWorkflowSteps'

export const StackStep = ({
  pollInterval = 5000,
  step,
  shouldPoll = false,
}: IPollStepDetails) => {
  const isGenerateStep = step?.name === 'generate install stack'
  const { org } = useOrg()
  const { install } = useInstall()
  const {
    data: stack,
    isLoading,
    error,
  } = usePolling<TInstallStack>({
    initIsLoading: true,
    path: `/api/${org.id}/installs/${step?.owner_id}/stack`,
    pollInterval,
    shouldPoll,
  })

  return (
    <>
      {isLoading && !stack ? (
        <div className="border rounded-md p-6">
          <Loading loadingText="Loading stack details..." variant="stack" />
        </div>
      ) : (
        <>
          {error?.error ? <Notice>{error?.error}</Notice> : null}
          {stack ? (
            isGenerateStep ? (
              <GenerateStack stack={stack} />
            ) : install?.app_runner_config.app_runner_type === 'aws' ? (
              <AwaitStack stack={stack} />
            ) : (
              <AwaitAzureStack stack={stack} />
            )
          ) : null}
        </>
      )}
    </>
  )
}

const GenerateStack = ({ stack }: { stack: TInstallStack }) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const {
    data: stackConfig,
    isLoading,
    error,
  } = useQuery<TAppStackConfig>({
    dependencies: [stack],
    path: `/api/${org.id}/apps/${install?.app_id}/configs/${
      stack?.versions?.at(0).app_config_id
    }`,
  })

  return (
    <>
      {isLoading && !stackConfig ? (
        <Loading loadingText="Loading stack infromation..." variant="stack" />
      ) : (
        <>
          {error?.error ? (
            <Notice>{error?.error || 'Unable to load stack config'}</Notice>
          ) : null}
          {stack ? (
            <>
              <>
                {stackConfig ? (
                  <div className="border p-3 rounded-md shadow flex flex-col gap-2">
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

const AwaitStack = ({ stack }: { stack: TInstallStack }) => {
  return (
    <>
      <div className="border rounded-md shadow flex flex-col">
        <div className="flex justify-between items-center p-3 border-b">
          <Text variant="med-14">
            Install stack{' '}
            {stack?.versions?.at(0)?.composite_status?.status === 'active'
              ? 'up and running'
              : 'is waiting to run'}
          </Text>
        </div>
        <div className="p-6 grid grid-cols-4">
          <StatusBadge
            status={stack?.versions?.at(0)?.composite_status?.status}
            description={
              stack?.versions?.at(0)?.composite_status?.status_human_description
            }
            label="Current status"
          />
          <div className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Last checked
            </Text>
            <Time
              time={stack?.versions?.at(0).runs?.at(-1)?.updated_at}
              format="relative"
            />
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-2">
        <Text variant="med-14">Setup your install stack</Text>
        <div className="border rounded-md shadow p-2 flex flex-col gap-1">
          <span className="flex justify-between items-center">
            <Text variant="med-12">Install quick link</Text>
            <ClickToCopyButton
              textToCopy={stack?.versions?.at(0)?.quick_link_url}
            />
          </span>
          <Link
            href={stack?.versions?.at(0)?.quick_link_url}
            target="_blank"
            rel="noopener noreferrer"
          >
            <Code>{stack?.versions?.at(0)?.quick_link_url}</Code>
          </Link>
        </div>

        <div className="border rounded-md shadow p-2 flex flex-col gap-1 mt-3">
          <span className="flex justify-between items-center">
            <Text variant="med-12">Install template link</Text>
            <ClickToCopyButton
              textToCopy={stack?.versions?.at(0)?.template_url}
            />
          </span>
          <Link
            href={stack?.versions?.at(0)?.template_url}
            target="_blank"
            rel="noopener noreferrer"
          >
            <Code>{stack?.versions?.at(0)?.template_url}</Code>
          </Link>
        </div>
      </div>

      <div className="relative">
        <hr />
        <Text className="shadow-sm px-2 border w-fit rounded-lg bg-white text-cool-grey-950 dark:bg-dark-grey-100 dark:text-cool-grey-50 absolute inset-0 m-auto h-[18px]">
          OR
        </Text>
      </div>

      <div className="flex flex-col gap-2">
        <Text variant="med-14">Setup your install stack using CLI command</Text>
        <div className="border rounded-md shadow p-2 flex flex-col gap-1">
          <ClickToCopyButton
            className="w-fit self-end"
            textToCopy={` aws cloudformation create-stack --stack-name [YOUR_STACK_NAME]
            --template-url ${stack?.versions?.at(0)?.template_url}`}
          />
          <Code>
            aws cloudformation create-stack --stack-name [YOUR_STACK_NAME]
            --template-url {stack?.versions?.at(0)?.template_url}
          </Code>
        </div>
      </div>

      <div className="flex flex-col gap-2">
        <Text variant="med-14">
          Update an existing install stack using CLI command
        </Text>
        <div className="border rounded-md shadow p-2 flex flex-col gap-1">
          <ClickToCopyButton
            className="w-fit self-end"
            textToCopy={` aws cloudformation update-stack --stack-name [YOUR_STACK_NAME]
            --template-url ${stack?.versions?.at(0)?.template_url}`}
          />
          <Code>
            aws cloudformation update-stack --stack-name [YOUR_STACK_NAME]
            --template-url {stack?.versions?.at(0)?.template_url}
          </Code>
        </div>
      </div>

      <div className="border p-3 rounded-md shadow flex flex-col gap-2">
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

const AwaitAzureStack = ({ stack }: { stack: TInstallStack }) => {
  const { install } = useInstall()
  return (
    <>
      <div className="border rounded-md shadow flex flex-col">
        <div className="flex justify-between items-center p-3 border-b">
          <Text variant="med-14">
            Install stack{' '}
            {stack?.versions?.at(0)?.composite_status?.status === 'active'
              ? 'up and running'
              : 'is waiting to run'}
          </Text>
        </div>
        <div className="p-6 grid grid-cols-4">
          <StatusBadge
            status={stack?.versions?.at(0)?.composite_status?.status}
            description={
              stack?.versions?.at(0)?.composite_status?.status_human_description
            }
            label="Current status"
          />
          <div className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Last checked
            </Text>
            <Time
              time={stack?.versions?.at(0).runs?.at(-1)?.updated_at}
              format="relative"
            />
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-2">
        <Text variant="med-14">
          Provision the install stack using the Azure CLI
        </Text>

        <div className="border rounded-md shadow p-2 flex flex-col gap-1">
          <span className="flex justify-between items-center">
            <Text variant="med-12">
              Ensure you are logged into the Azure subscription you want to
              install into
            </Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={`az login`}
            />
          </span>
          <Code>az login</Code>
        </div>

        <div className="border rounded-md shadow p-2 flex flex-col gap-1">
          <span className="flex justify-between items-center">
            <Text variant="med-12">Create a resource group to deploy into</Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={`az group create --name ${install.id}-rg --location ${install.azure_account.location}`}
            />
          </span>
          <Code>{`
            az group create --name ${install.id}-rg --location ${install.azure_account.location}
          `}</Code>
        </div>
      </div>

      <div className="flex flex-col gap-2">
        <div className="border rounded-md shadow p-2 flex flex-col gap-1">
          <span className="flex justify-between items-center">
            <Text variant="med-12">Deploy the stack to the resource group</Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={`az stack group create --name ${install.id}-stack --resource-group ${install.id}-rg --template-uri ${stack?.versions?.at(0)?.template_url} --deny-settings-mode "denyDelete" --aou deleteAll`}
            />
          </span>
          <Code>{`
            az stack group create --name ${install.id}-stack --resource-group ${install.id}-rg --template-uri ${stack?.versions?.at(0)?.template_url} --deny-settings-mode "denyDelete" --aou deleteAll
          `}</Code>
        </div>
      </div>

      <div className="relative">
        <hr />
        <Text className="shadow-sm px-2 border w-fit rounded-lg bg-white text-cool-grey-950 dark:bg-dark-grey-100 dark:text-cool-grey-50 absolute inset-0 m-auto h-[18px]">
          OR
        </Text>
      </div>

      <div className="flex flex-col gap-2">
        <Text variant="med-14">Download the install stack template</Text>
        <div className="border rounded-md shadow p-2 flex flex-col gap-1 mt-3">
          <span className="flex justify-between items-center">
            <Text variant="med-12">Install template link</Text>
            <ClickToCopyButton
              textToCopy={stack?.versions?.at(0)?.template_url}
            />
          </span>
          <Link
            href={stack?.versions?.at(0)?.template_url}
            target="_blank"
            rel="noopener noreferrer"
          >
            <Code>{stack?.versions?.at(0)?.template_url}</Code>
          </Link>
        </div>
      </div>

      <div className="border p-3 rounded-md shadow flex flex-col gap-2">
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
