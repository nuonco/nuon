import { useOutletContext, useParams } from 'react-router'
import { PageTitle } from '@/components/navigation/PageTitle'
import { Text } from '@/components/common/Text'
import { RunbookStep } from '@/components/runbooks/RunbookStep'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import type { TRunbookOutletContext } from './types'

export const RunbookStepsTab = () => {
  const { runbook } = useOutletContext<TRunbookOutletContext>()
  const { org } = useOrg()
  const { app } = useApp()
  const { branchId } = useParams()
  const actionBasePath = branchId
    ? `/${org?.id}/apps/${app?.id}/branches/${branchId}`
    : `/${org?.id}/apps/${app?.id}`

  const latestConfig = runbook?.configs?.[0]
  const steps =
    latestConfig?.steps?.slice().sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0)) ??
    []

  return (
    <>
      <PageTitle
        segments={[`${runbook?.name ?? 'Runbook'} steps`, app?.name]}
      />
      {!steps.length ? (
        <Text theme="neutral">No steps configured.</Text>
      ) : (
        <div className="grid grid-cols-1 gap-4">
          {steps.map((step, i) => (
            <RunbookStep
              key={step.id ?? i}
              index={i}
              step={step}
              actionBasePath={actionBasePath}
            />
          ))}
        </div>
      )}
    </>
  )
}
