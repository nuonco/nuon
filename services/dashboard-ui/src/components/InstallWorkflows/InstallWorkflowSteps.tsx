import React, { type FC } from 'react'
import { Empty } from '@/components/Empty'
import { Expand } from '@/components/Expand'
import { Link } from '@/components/Link'
import { Notice } from '@/components/Notice'
import { Duration } from '@/components/Time'
import { Text, Code } from '@/components/Typography'
import type { TInstallWorkflow, TInstallWorkflowStep } from '@/types'
import { sentanceCase } from '@/utils'
import { YAStatus } from './InstallWorkflowHistory'

function buildDetailsHref(
  step: TInstallWorkflowStep,
  orgId: string
): string | null {
  const basePath = `/${orgId}/installs/${step?.install_id}`

  let href: string | null = null
  switch (step?.step_target_type) {
    case 'install_action_workflow_runs':
      break

    case 'install_deploys':
      break

    case 'install_sandbox_runs':
      href = `${basePath}/sandbox/${step?.step_target_id}`
      break
    default:
      href = null
  }

  return href
}

interface IInstallWorkflowSteps {
  installWorkflow: TInstallWorkflow
  orgId: string
}

export const InstallWorkflowSteps: FC<IInstallWorkflowSteps> = ({
  installWorkflow,
  orgId,
}) => {
  return (
    <div className="flex flex-col gap-2">
      {installWorkflow?.steps?.length ? (
        installWorkflow?.steps?.map((step, i) => {
          const href = buildDetailsHref(step, orgId)
          return (
            <Expand
              key={step?.id}
              heading={
                <InstallWorkflowStepTitle
                  executionTime={step?.execution_time}
                  name={step?.name}
                  status={step?.status}
                  stepNumber={i + 1}
                />
              }
              hasHeadingStyle
              headerClass="p-3 !pr-3 !border-none"
              className="w-full"
              id={step?.id}
              parentClass="border rounded-md overflow-hidden"
              expandContent={
                <div className="p-3 border-t flex flex-col gap-4">
                  {step?.status?.metadata?.reason ? (
                    <Notice variant="warn">
                      {sentanceCase(step?.status?.metadata?.reason as string)}
                    </Notice>
                  ) : null}
                  {step?.status?.metadata?.err_message ? (
                    <Notice variant="error" className="!items-start">
                      {step?.status?.metadata?.err_message as string}
                    </Notice>
                  ) : null}
                  {href ? (
                    <Link className="text-sm" href={href}>
                      View details
                    </Link>
                  ) : null}
                  <div>
                    <Text variant="med-12" className="mb-2">
                      Step JSON
                    </Text>
                    <Code variant="preformated">
                      {JSON.stringify(step, null, 2)}
                    </Code>
                  </div>
                </div>
              }
            />
          )
        })
      ) : (
        <Empty
          emptyTitle="Waiting on steps"
          emptyMessage="Waiting on update steps to generate."
          variant="history"
        />
      )}
    </div>
  )
}

const InstallWorkflowStepTitle: FC<{
  executionTime: number
  name: string
  status: TInstallWorkflowStep['status']
  stepNumber: number
}> = ({ executionTime, name, status, stepNumber }) => {
  return (
    <span className="flex justify-between w-full pr-3">
      <span className="flex gap-2">
        <YAStatus status={status?.status} />
        <span>
          <Text className="text-cool-grey-600 dark:text-white/70">
            Step {stepNumber}
          </Text>
          <Text variant="reg-14">{sentanceCase(name)}</Text>
        </span>
      </span>

      {status?.status === 'active' ||
      status?.status === 'error' ||
      status?.status === 'success' ? (
        <Text
          variant="reg-12"
          className="text-cool-grey-600 dark:text-white/70"
        >
          Executed in <Duration nanoseconds={executionTime} />
        </Text>
      ) : null}
    </span>
  )
}
