import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { DebouncedSearchInput } from '@/components/common/DeboundedSearch'
import { LatestRunbookRunCard } from '@/components/runbooks/LatestRunbookRunCard'
import { InstallRunbookRowActions } from '@/components/runbooks/InstallRunbookRowActions'
import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
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
  const installLink = useInstallLink()
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
    const href = installLink({
      installId: install?.id,
      appId: install?.app_id,
      suffix: `/runbooks/${runbookId}`,
    })
    const latestRun = installRunbook.runs?.[0]
    const workflowId =
      latestRun?.install_workflow_id ?? latestRun?.install_workflow?.id
    const runHref = workflowId
      ? installLink({
          installId: install?.id,
          appId: install?.app_id,
          suffix: `/history/${workflowId}`,
        })
      : undefined

    return {
      id: runbookId,
      name: runbook?.name ?? 'Runbook',
      href,
      description: runbook?.description,
      stepCount: runbook?.configs?.[0]?.steps?.length,
      actions: <InstallRunbookRowActions installRunbook={installRunbook} />,
      latestRun: <LatestRunbookRunCard flush run={latestRun} href={runHref} />,
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
