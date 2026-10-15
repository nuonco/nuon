import { Expand } from '@/components/common/Expand'
import { JSONViewer } from '@/components/common/JSONViewer'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Text } from '@/components/common/Text'
import { CompositeError } from '@/components/common/CompositeError'
import type { TInstallActionRun, TAccount, TWorkflowStep } from '@/types'
import { ActionRunHeader } from '../ActionRunHeader'
import { ActionRunMetadata } from '../ActionRunMetadata'
import { AdhocActionDetails } from '../AdhocActionDetails'
import { StandardActionSteps } from '../StandardActionSteps'
import { ActionRunLogs } from '../ActionRunLogs'

export interface IActionRunStepDetails {
  step?: TWorkflowStep
  actionRun?: TInstallActionRun
  createdBy?: TAccount
  error: any
  isLoading: boolean
}

export const ActionRunStepDetails = ({
  step,
  actionRun,
  createdBy,
  error,
  isLoading,
}: IActionRunStepDetails) => {
  const isAdhoc = actionRun?.trigger_type === 'adhoc'
  const envVarEntries = Object.entries({
    ...actionRun?.steps?.at(0)?.adhoc_config?.env_vars,
    ...actionRun?.run_env_vars,
  })
  const stepOutputs = actionRun?.outputs?.steps ?? {}
  const outputSteps = Object.keys(stepOutputs)
  const outputs =
    outputSteps.length === 1 ? stepOutputs[outputSteps[0]] : stepOutputs
  const outputCount = Object.keys(outputs ?? {}).length

  if (isLoading && !actionRun) {
    return (
      <div className="flex flex-col gap-4">
        <ActionRunHeader loading />
        <ActionRunMetadata loading />
        <StandardActionSteps loading />
        <ActionRunLogs loading />
      </div>
    )
  }

  if (error || !actionRun) {
    return (
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong" theme="error">
          Unable to load action run details
        </Text>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <ActionRunHeader actionRun={actionRun} isAdhoc={isAdhoc} step={step} />

      {actionRun?.composite_error ? (
        <CompositeError error={actionRun.composite_error} />
      ) : null}

      <ActionRunMetadata
        actionRun={actionRun}
        createdBy={createdBy}
        step={step}
      />

      {isAdhoc ? (
        <AdhocActionDetails actionRun={actionRun} />
      ) : (
        <StandardActionSteps actionRun={actionRun} />
      )}

      <Expand
        id={`${actionRun.id}-outputs`}
        className="rounded-md border"
        heading={<SectionHeading title="Outputs" count={outputCount} />}
      >
        <div className="border-t">
          {outputCount ? (
            <JSONViewer
              className="border-none !rounded-t-none max-h-[600px] overflow-auto"
              data={outputs}
            />
          ) : (
            <Text variant="subtext" theme="neutral" className="p-4">
              No outputs
            </Text>
          )}
        </div>
      </Expand>

      {envVarEntries.length > 0 ? (
        <Expand
          id={`${actionRun.id}-env-vars`}
          className="rounded-md border"
          heading={
            <SectionHeading
              title="Environment variables"
              count={envVarEntries.length}
            />
          }
        >
          <div className="border-t p-4">
            <KeyValueList
              values={envVarEntries.map(([key, value]) => ({ key, value }))}
            />
          </div>
        </Expand>
      ) : null}

      <ActionRunLogs actionRun={actionRun} isAdhoc={isAdhoc} step={step} />
    </div>
  )
}

const SectionHeading = ({ title, count }: { title: string; count: number }) => (
  <span className="flex items-center gap-2">
    <Text weight="strong">{title}</Text>
    <Text variant="subtext" theme="neutral">
      {count}
    </Text>
  </span>
)
