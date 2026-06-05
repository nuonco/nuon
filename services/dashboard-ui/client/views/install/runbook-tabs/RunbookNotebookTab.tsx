import { useOutletContext } from 'react-router'
import { Markdown } from '@/components/common/Markdown'
import { Text } from '@/components/common/Text'
import { RunbookStep } from '@/components/runbooks/RunbookStep'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import type { TInstallRunbookOutletContext } from './types'

export const RunbookNotebookTab = () => {
  const { installRunbook } = useOutletContext<TInstallRunbookOutletContext>()
  const { org } = useOrg()
  const { install } = useInstall()

  const latestConfig = installRunbook?.runbook?.configs?.[0]
  const cells = latestConfig?.cells ?? []
  const steps =
    latestConfig?.steps
      ?.slice()
      .sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0)) ?? []

  if (!cells.length) {
    return <Text theme="neutral">No notebook cells configured.</Text>
  }

  return (
    <div className="flex flex-col gap-6">
      {cells.map((cell, i) => {
        if (cell.type === 'markdown') {
          return (
            <Markdown
              key={`cell-${i}`}
              content={cell.content ?? ''}
              mode="install"
            />
          )
        }

        const step =
          cell.step_idx != null ? steps[cell.step_idx] : undefined
        if (!step) return null

        return (
          <RunbookStep
            key={`cell-${i}`}
            index={cell.step_idx ?? i}
            step={step}
            actionBasePath={`/${org?.id}/installs/${install?.id}`}
          />
        )
      })}
    </div>
  )
}
