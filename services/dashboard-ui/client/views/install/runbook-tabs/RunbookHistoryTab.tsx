import { useOutletContext } from 'react-router'
import { RunbookRunTimeline } from '@/components/runbooks/RunbookRunTimeline'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import type { TInstallRunbookOutletContext } from './types'

export const RunbookHistoryTab = () => {
  const { installRunbook } = useOutletContext<TInstallRunbookOutletContext>()
  const { org } = useOrg()
  const { install } = useInstall()

  const runbook = installRunbook?.runbook
  const runs = installRunbook?.runs ?? []

  return (
    <RunbookRunTimeline
      runs={runs}
      runbookName={runbook?.name ?? ''}
      basePath={`/${org?.id}/installs/${install?.id}`}
    />
  )
}
