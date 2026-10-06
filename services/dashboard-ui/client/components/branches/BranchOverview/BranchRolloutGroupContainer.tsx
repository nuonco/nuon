import { useCallback, useEffect, useMemo, useRef } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router'
import { PlanDiffPanel } from '@/components/branches/BranchRunApproval/PlanDiffPanel'
import { PlanGroupStep } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep'
import { Button } from '@/components/common/Button'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { SkipStepButton } from '@/components/workflows/step-details/SkipStep/SkipStepContainer'
import type { TAppBranchRun, TWorkflowStep } from '@/types'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useSurfaces } from '@/hooks/use-surfaces'
import { InstallFailureNotice } from './BranchOverview'
import { InstallRolloutPanel } from './InstallRolloutPanel'
import { installFailureHref, overviewCompositeError } from './overview-loading'
import { RolloutGroupDetail } from './RolloutGroupDetail'
import { deployStepForGroup, planStepForGroup } from './rollout-stages'
import { useRolloutGroups } from './use-rollout-groups'

export const BranchRolloutGroupContainer = () => {
  const params = useParams()
  const groupId = params.groupId
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const installId = searchParams.get('install') ?? undefined
  const { addPanel, updatePanel, removePanel, panels } = useSurfaces()
  const panelIdRef = useRef<string | null>(null)
  const openedFor = useRef<string | null>(null)
  const openPanelId =
    panels.find((panel) => panel?.id === panelIdRef.current)?.id ?? null
  const {
    app,
    orgId,
    repoSlug,
    rolloutHref,
    rollout,
    branchRun,
    workflowSteps,
    groups,
    hasPlan,
    isLoading,
  } = useRolloutGroups()

  const group = groups.find((item) => item.id === groupId)

  useEffect(() => {
    if (isLoading || !groups.length || group || !groupId) return
    navigate(rolloutHref, { replace: true })
  }, [isLoading, groups.length, group, groupId, navigate, rolloutHref])

  const planStep = group
    ? planStepForGroup(workflowSteps, group.name)
    : undefined
  const deployStep = group
    ? deployStepForGroup(workflowSteps, group.name)
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
    const found = (group?.installs ?? []).find(
      (install) => install.id === installId
    )
    if (found) return found
    if (!installId) return undefined
    return { id: installId, name: installId, status: 'pending' }
  }, [group?.installs, installId])

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

  const selectInstall = (_nextGroupId: string, nextInstallId: string) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
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
        approval={group?.approval}
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
    group?.approval,
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
      <PageTitle
        segments={[group?.name ?? 'Group', app?.name]}
      />
      <SectionHeader
        title={group?.name ?? 'Group'}
        description={rollout?.activity}
        actions={
          <span className="flex items-center gap-3">
            {planStep?.id && rollout?.id ? (
              <Button
                size="sm"
                variant="secondary"
                onClick={() => {
                  if (!planStep) return
                  addPanel(
                    <PlanDiffPanel
                      step={planStep}
                      installFacts={
                        group
                          ? Object.fromEntries(
                              group.installs.map((install) => [
                                install.id,
                                {
                                  labels: install.labels,
                                  region: install.region,
                                  status: install.status,
                                  detail: install.detail,
                                  appliedConfigId: install.appliedConfigId,
                                },
                              ])
                            )
                          : undefined
                      }
                    />,
                    `plan-diff-${planStep.id}`
                  )
                }}
              >
                View plan
              </Button>
            ) : null}
            <Link href={rolloutHref}>Back to rollout</Link>
          </span>
        }
      />
      {isLoading ? (
        <Loading />
      ) : !hasPlan ? (
        <Text variant="subtext" theme="neutral">
          This branch has no install groups yet.
        </Text>
      ) : !group ? (
        <Text variant="subtext" theme="neutral">
          This group is not part of the current rollout.{' '}
          <Link href={rolloutHref}>Back to rollout</Link>
        </Text>
      ) : (
        <RolloutGroupDetail
          group={group}
          onSelectInstall={selectInstall}
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
                  group.installs.map((install) => [
                    install.id,
                    {
                      labels: install.labels,
                      region: install.region,
                      status: install.status,
                      detail: install.detail,
                      appliedConfigId: install.appliedConfigId,
                    },
                  ])
                )}
                onSelectInstall={(id) => selectInstall(group.id, id)}
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
