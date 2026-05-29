import { useOutletContext } from 'react-router'
import { Text } from '@/components/common/Text'
import { RunbookStep } from '@/components/runbooks/RunbookStep'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import type { TRunbookOutletContext } from './types'

export const RunbookStepsTab = () => {
  const { runbook } = useOutletContext<TRunbookOutletContext>()
  const { org } = useOrg()
  const { app } = useApp()

  const latestConfig = runbook?.configs?.[0]
  const steps =
    latestConfig?.steps
      ?.slice()
      .sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0)) ?? []

  if (!steps.length) {
    return <Text theme="neutral">No steps configured.</Text>
  }

  return (
    <div className="grid grid-cols-1 gap-4">
      {steps.map((step, i) => (
        <RunbookStep
          key={step.id ?? i}
          index={i}
          step={step}
          actionBasePath={`/${org?.id}/apps/${app?.id}`}
        />
      ))}
    </div>
  )
}
