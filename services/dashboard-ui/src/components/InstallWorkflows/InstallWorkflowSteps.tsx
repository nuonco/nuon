import React, { type FC } from 'react'
import { Empty } from '@/components/Empty'
import { Expand } from '@/components/Expand'
import { Text, Code } from '@/components/Typography'
import type { TInstallWorkflow, TInstallWorkflowStep } from '@/types'
import { sentanceCase } from '@/utils'
import { YAStatus } from './InstallWorkflowHistory'

interface IInstallWorkflowSteps {
  installWorkflow: TInstallWorkflow
}

export const InstallWorkflowSteps: FC<IInstallWorkflowSteps> = ({
  installWorkflow,
}) => {
  return (
    <div className="flex flex-col gap-2">
      {installWorkflow?.steps?.length ? (
        installWorkflow?.steps?.map((step, i) => (
          <Expand
            key={step?.id}
            heading={
              <InstallWorkflowStepTitle
                name={step?.name}
                status={step?.status}
                stepNumber={i + 1}
              />
            }
            hasHeadingStyle
            headerClass="p-3 !pr-3 !border-none"
            id={step?.id}
            parentClass="border rounded-md overflow-hidden"
            expandContent={
              <div className="p-3 border-t">
                <Code variant="preformated">
                  {JSON.stringify(step?.status, null, 2)}
                </Code>
              </div>
            }
          />
        ))
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
  name: string
  status: TInstallWorkflowStep['status']
  stepNumber: number
}> = ({ status, name, stepNumber }) => {
  return (
    <span className="flex gap-2">
      <YAStatus status={status?.status} />
      <span>
        <Text className="text-cool-grey-600 dark:text-white/70">
          Step {stepNumber}
        </Text>
        <Text variant="reg-14">{sentanceCase(name)}</Text>
      </span>
    </span>
  )
}
