import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { DebouncedSearchInput } from '@/components/common/DeboundedSearch'
import { LatestRunbookRunCard } from '@/components/runbooks/LatestRunbookRunCard'
import { RunbookReadmePanel } from '@/components/runbooks/RunbookReadmePanel'
import { RunRunbookButton } from '@/components/runbooks/RunRunbook'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallRunbooks } from '@/lib'
import type { TInstallRunbook } from '@/lib/ctl-api/installs/runbooks'
import {
  InstallRunbooksList,
  type TInstallRunbookListItem,
} from './InstallRunbooksList'

const LIMIT = 10

export const InstallRunbooksListContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)
  const q = searchParams.get('q') || undefined

  const { data: result, isLoading } = useQuery({
    queryKey: ['install-runbooks', org?.id, install?.id, offset, q],
    queryFn: () =>
      getInstallRunbooks({
        orgId: org!.id,
        installId: install!.id,
        offset,
        limit: LIMIT,
        q,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: 20000,
    enabled: !!org?.id && !!install?.id,
  })

  const toListItem = (
    installRunbook: TInstallRunbook
  ): TInstallRunbookListItem => {
    const runbook = installRunbook.runbook
    const runbookId = installRunbook.runbook_id ?? installRunbook.id
    const href = `/${org?.id}/installs/${install?.id}/runbooks/${runbookId}`
    const readme = runbook?.configs?.[0]?.readme
    const latestRun = installRunbook.runs?.[0]
    const workflowId =
      latestRun?.install_workflow_id ?? latestRun?.install_workflow?.id
    const runHref = workflowId
      ? `/${org?.id}/installs/${install?.id}/history/${workflowId}`
      : undefined

    return {
      id: runbookId,
      name: runbook?.name ?? 'Runbook',
      href,
      description: runbook?.description,
      stepCount: runbook?.configs?.[0]?.steps?.length,
      actions: (
        <div className="flex items-center gap-2">
          {readme ? (
            <RunbookReadmePanel
              readme={readme}
              runbookName={runbook?.name}
              panelKey={`runbook-readme-${runbookId}`}
              triggerButton={{
                variant: 'secondary',
                size: 'sm',
                children: 'Readme',
              }}
            />
          ) : null}
          <RunRunbookButton
            installRunbook={installRunbook}
            size="sm"
            variant="secondary"
          >
            Run runbook
          </RunRunbookButton>
        </div>
      ),
      latestRun: (
        <LatestRunbookRunCard flush run={latestRun} href={runHref} />
      ),
    }
  }

  return (
    <InstallRunbooksList
      items={(result?.data ?? []).map(toListItem)}
      loading={isLoading}
      filtered={!!q}
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
      search={
        <DebouncedSearchInput
          className="w-full md:w-fit"
          labelClassName="w-full md:w-fit"
          placeholder="Search by name or ID..."
        />
      }
    />
  )
}
