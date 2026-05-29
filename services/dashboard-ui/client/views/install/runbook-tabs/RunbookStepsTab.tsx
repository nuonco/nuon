import { useOutletContext } from 'react-router'
import { Text } from '@/components/common/Text'
import { RunbookStep } from '@/components/runbooks/RunbookStep'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import type { TInstallRunbookOutletContext } from './types'

export const RunbookStepsTab = () => {
  const { installRunbook } = useOutletContext<TInstallRunbookOutletContext>()
  const { org } = useOrg()
  const { install } = useInstall()

  const latestConfig = installRunbook?.runbook?.configs?.[0]
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
          actionBasePath={`/${org?.id}/installs/${install?.id}`}
        />
      ))}
    </div>
  )
}
