import React, { type FC } from 'react'
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
      {installWorkflow?.steps?.map((step) => (
        <Expand
          key={step?.id}
          heading={
            <InstallWorkflowStepTitle name={step?.name} status={step?.status} />
          }
          headerClass="p-3 !pr-3"
          id={step?.id}
          parentClass="border rounded-md"
          expandContent={
            <div className="p-3 border-t">
              <Code variant="preformated">
                {JSON.stringify(step?.status, null, 2)}
              </Code>
            </div>
          }
        />
      ))}
    </div>
  )
}

const InstallWorkflowStepTitle: FC<{
  name: string
  status: TInstallWorkflowStep['status']
}> = ({ status, name }) => {
  return (
    <span className="flex gap-2">
      <YAStatus status={status?.status} />
      <Text variant="reg-14">{sentanceCase(name)}</Text>
    </span>
  )
}
