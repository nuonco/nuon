import { useCallback, useEffect, useMemo, useRef } from 'react'
import { useSearchParams } from 'react-router'
import { PlanGroupStep } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep'
import { CommitLink } from '@/components/common/GitReferenceLink'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { SkipStepButton } from '@/components/workflows/step-details/SkipStep/SkipStepContainer'
import type { TAppBranchRun, TWorkflowStep } from '@/types'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useSurfaces } from '@/hooks/use-surfaces'
import { InstallFailureNotice, type TOverviewRollout } from './BranchOverview'
import { InstallRolloutPanel } from './InstallRolloutPanel'
import { installFailureHref, overviewCompositeError } from './overview-loading'
import { deployStepForGroup, planStepForGroup } from './rollout-stages'
import { RolloutTrack, selectedTrackGroup } from './RolloutTrack'
import { useRolloutGroups } from './use-rollout-groups'

const RolloutRunCard = ({ rollout }: { rollout: TOverviewRollout }) => {
  const message = rollout.commit?.message?.split('\n')[0]
  const sha = rollout.commit?.sha ?? rollout.sha
  const shaUrl = rollout.commit?.shaUrl ?? rollout.shaUrl
  const author = rollout.commit?.author ?? rollout.author
  return (
    <div className="flex flex-col gap-2 rounded-xl border bg-white px-4 py-3 shadow-sm dark:bg-dark-grey-900">
      <Text variant="body" weight="strong" className="break-words">
        {message || rollout.title}
      </Text>
      {sha || author ? (
        <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
          {sha ? <CommitLink sha={sha} href={shaUrl} /> : null}
          {author ? (
            <Text variant="subtext" theme="neutral">
              {author}
            </Text>
          ) : null}
        </span>
      ) : null}
    </div>
  )
}

export const BranchRolloutContainer = () => {
  const [searchParams, setSearchParams] = useSearchParams()
  const groupId = searchParams.get('group') ?? undefined
  const installId = searchParams.get('install') ?? undefined
  const { addPanel, updatePanel, removePanel, panels } = useSurfaces()
  const panelIdRef = useRef<string | null>(null)
  const openedFor = useRef<string | null>(null)
  const openPanelId =
    panels.find((panel) => panel?.id === panelIdRef.current)?.id ?? null
  const {
    app,
    branch,
    orgId,
    repoSlug,
    basePath,
    rollout,
    branchRun,
    workflowSteps,
    groups,
    hasPlan,
    isLoading,
  } = useRolloutGroups()

  const selected = selectedTrackGroup(groups, groupId)
  const planStep = selected
    ? planStepForGroup(workflowSteps, selected.name)
    : undefined
  const deployStep = selected
    ? deployStepForGroup(workflowSteps, selected.name)
    : undefined
  const canSkipDeploy =
    !!deployStep?.skippable &&
    deployStep.status?.status === 'error' &&
    !!deployStep.install_workflow_id
  const compositeError = overviewCompositeError(
    workflowSteps,
    branchRun?.composite_error
  )
  const selectedInstall = useMemo(() => {
    const found = groups
      .flatMap((group) => group.installs)
      .find((install) => install.id === installId)
    if (found) return found
    if (!installId) return undefined
    return { id: installId, name: installId, status: 'pending' }
  }, [groups, installId])

  const closeInstall = useCallback(() => {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        next.delete('install')
        next.delete('plan')
        return next
      },
      { replace: true }
    )
  }, [setSearchParams])

  const selectGroup = (id: string) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      next.set('group', id)
      next.delete('install')
      next.delete('plan')
      return next
    })
  }

  const selectInstall = (nextGroupId: string, nextInstallId: string) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      next.set('group', nextGroupId)
      next.set('install', nextInstallId)
      next.delete('plan')
      return next
    })
  }

  useEffect(() => {
    if (!installId || !selectedInstall) {
      if (openPanelId) {
        removePanel(openPanelId)
        panelIdRef.current = null
        openedFor.current = null
      }
      return
    }
    if (openPanelId && openedFor.current !== installId) {
      removePanel(openPanelId)
      panelIdRef.current = null
      openedFor.current = null
      return
    }
    const panel = (
      <InstallRolloutPanel
        install={selectedInstall}
        approval={selected?.approval}
        orgId={orgId}
        repo={repoSlug}
        branchRun={branchRun as TAppBranchRun | undefined}
        onClose={closeInstall}
      />
    )
    if (openPanelId) {
      updatePanel(openPanelId, panel)
      return
    }
    openedFor.current = installId
    panelIdRef.current = addPanel(panel)
  }, [
    installId,
    selectedInstall,
    selected?.approval,
    orgId,
    repoSlug,
    branchRun,
    openPanelId,
    addPanel,
    updatePanel,
    removePanel,
    closeInstall,
  ])

  return (
    <div className="flex flex-col gap-3 p-4 md:p-6">
      <PageTitle segments={['Rollout', branch?.name, app?.name]} />
      <SectionHeader
        title="Rollout"
        description={rollout?.activity}
        actions={
          rollout ? (
            <span className="flex items-center gap-3">
              <Status status={rollout.status} />
              <Link href={rollout.href}>View run</Link>
            </span>
          ) : undefined
        }
      />
      {rollout ? <RolloutRunCard rollout={rollout} /> : null}
      {isLoading ? (
        <Loading />
      ) : !hasPlan ? (
        <Text variant="subtext" theme="neutral">
          This branch has no install groups yet. Every install updates at once.{' '}
          <Link href={`${basePath}/settings`}>Create a deployment plan</Link>
        </Text>
      ) : (
        <RolloutTrack
          groups={groups}
          selectedGroupId={groupId}
          onSelectGroup={selectGroup}
          onSelectInstall={selectInstall}
          emptyMessage="No runs yet. Push a commit or start a run."
          notice={
            compositeError ? (
              <InstallFailureNotice
                error={compositeError}
                href={installFailureHref(compositeError, orgId)}
              />
            ) : undefined
          }
          summary={
            planStep ? (
              <PlanGroupStep
                step={planStep}
                metadata={
                  (planStep.status?.metadata ?? {}) as Record<string, any>
                }
                workflowStatus={rollout?.status}
                hideHeading
                installFacts={Object.fromEntries(
                  (selected?.installs ?? []).map((install) => [
                    install.id,
                    { labels: install.labels, region: install.region },
                  ])
                )}
                onSelectInstall={(id) =>
                  selected ? selectInstall(selected.id, id) : undefined
                }
              />
            ) : undefined
          }
          footer={
            canSkipDeploy && deployStep ? (
              <div className="flex flex-col items-start gap-2">
                <Text variant="subtext" theme="neutral">
                  Skip this group to continue the rollout.
                </Text>
                <SkipStepButton step={deployStep as TWorkflowStep}>
                  Skip and continue
                </SkipStepButton>
              </div>
            ) : undefined
          }
        />
      )}
    </div>
  )
}
