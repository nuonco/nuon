import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallRunbooks } from '@/lib'
import { RunRunbookModal } from '../RunRunbook/RunRunbook'
import { RunRunbookCard } from './RunRunbookCard'

interface IRunRunbookCardContainer {
  id?: string
  name?: string
}

export const RunRunbookCardContainer = ({ id, name }: IRunRunbookCardContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { addModal } = useSurfaces()

  const { data: result, isLoading, error } = useQuery({
    queryKey: ['install-runbooks-card', org?.id, install?.id, name, id],
    queryFn: () =>
      getInstallRunbooks({
        orgId: org!.id,
        installId: install!.id,
        limit: 50,
        offset: 0,
      }),
    enabled: !!org?.id && !!install?.id && !!(name || id),
  })

  if (!id && !name) {
    return <RunRunbookCard error="Missing id or name attribute" />
  }

  const runbooks = result?.data ?? []
  const runbook = name
    ? runbooks.find((r) => r.runbook?.name === name)
    : runbooks.find((r) => r.id === id || r.runbook_id === id)

  if (!isLoading && !error && runbooks.length > 0 && !runbook) {
    return <RunRunbookCard error={`Runbook "${name || id}" not found`} />
  }

  const steps = runbook?.runbook?.configs?.[0]?.steps ?? []
  const runbookId = runbook?.runbook_id ?? runbook?.id
  const href = runbookId
    ? `/${org?.id}/installs/${install?.id}/runbooks/${runbookId}`
    : undefined

  return (
    <RunRunbookCard
      name={runbook?.runbook?.name || name}
      href={href}
      stepCount={runbook ? steps.length : undefined}
      isLoading={isLoading}
      error={error ? 'Failed to load runbook' : undefined}
      onRun={() => {
        if (runbook) {
          addModal(<RunRunbookModal installRunbook={runbook} />)
        }
      }}
    />
  )
}
