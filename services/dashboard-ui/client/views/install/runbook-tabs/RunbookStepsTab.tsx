import { useOutletContext } from 'react-router'
import { PageTitle } from '@/components/navigation/PageTitle'
import { Text } from '@/components/common/Text'
import { RunbookStep } from '@/components/runbooks/RunbookStep'
import { useInstallPage } from '@/hooks/use-install-path'
import type { TInstallRunbookOutletContext } from './types'

export const RunbookStepsTab = () => {
  const { installRunbook } = useOutletContext<TInstallRunbookOutletContext>()
  const { install, href } = useInstallPage()

  const latestConfig = installRunbook?.runbook?.configs?.[0]
  const steps =
    latestConfig?.steps
      ?.slice()
      .sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0)) ?? []

  return (
    <>
      <PageTitle
        segments={[
          `${installRunbook?.runbook?.name ?? 'Runbook'} steps`,
          install?.name,
        ]}
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
              actionBasePath={href()}
            />
          ))}
        </div>
      )}
    </>
  )
}
