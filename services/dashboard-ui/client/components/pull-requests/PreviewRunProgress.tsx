import { useEffect, useState } from 'react'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { BuildRow } from '@/components/branches/WorkflowStepDetail/steps/BuildStep'
import type { TOverviewStage } from '@/components/branches/BranchOverview/overview-loading'
import type { TComponentType } from '@/types'

const FADE_MS = 200

export type TPreviewRunBuild = {
  component_id?: string
  component_name?: string
  component_type?: TComponentType | 'sandbox'
  status?: string
  change_reason?: string
  build_id?: string
}

export interface IPreviewRunProgress {
  stages: TOverviewStage[]
  description?: string
  builds?: TPreviewRunBuild[]
  sandboxBuildId?: string
  sha?: string
  baseRef?: string
  install?: string
  orgId?: string
  appId?: string
  variant?: 'page' | 'inline'
}

const activeStage = (stages: TOverviewStage[]) =>
  stages.find((stage) => stage.status === 'in-progress') ??
  stages.find((stage) => stage.status === 'pending') ??
  stages.at(-1)

const stageCopy = (
  id: string,
  {
    sha,
    baseRef,
    install,
  }: Pick<IPreviewRunProgress, 'sha' | 'baseRef' | 'install'>
) => {
  switch (id) {
    case 'starting':
      return {
        title: 'Starting the preview',
        body: 'Nuon is queuing a preview run for this pull request.',
      }
    case 'fetch-commit':
      return {
        title: 'Fetching the commit',
        body: sha
          ? `Reading ${sha.slice(0, 7)} from GitHub.`
          : 'Reading the latest commit on this pull request from GitHub.',
      }
    case 'app-config':
      return {
        title: 'Compiling the app config',
        body: `Building the config at this commit and comparing it with ${baseRef || 'the base branch'}.`,
      }
    case 'build-components':
      return {
        title: 'Building components',
        body: 'Rebuilding the components and sandbox this change touches.',
      }
    case 'plan':
      return {
        title: install ? `Planning against ${install}` : 'Planning the change',
        body: 'Running a plan to show what would change. Nothing is applied.',
      }
    case 'apply':
      return {
        title: install ? `Applying to ${install}` : 'Applying the change',
        body: 'Deploying this change to the preview install.',
      }
    default:
      return { title: 'Running the preview', body: '' }
  }
}

export const PreviewRunProgress = ({
  stages,
  description,
  builds = [],
  sandboxBuildId,
  sha,
  baseRef,
  install,
  orgId,
  appId,
  variant = 'page',
}: IPreviewRunProgress) => {
  const stage = activeStage(stages)
  const stageId = stage?.id ?? 'starting'
  const [shownId, setShownId] = useState(stageId)
  const [visible, setVisible] = useState(true)

  useEffect(() => {
    if (stageId === shownId) return
    setVisible(false)
    const swap = setTimeout(() => {
      setShownId(stageId)
      setVisible(true)
    }, FADE_MS)
    return () => clearTimeout(swap)
  }, [stageId, shownId])

  const copy = stageCopy(shownId, { sha, baseRef, install })
  const step = stages.findIndex((item) => item.id === shownId)
  const showBuilds = shownId === 'build-components' && builds.length > 0
  const detail =
    description &&
    shownId === stageId &&
    description.toLowerCase() !== copy.title.toLowerCase()
      ? description
      : undefined

  if (variant === 'inline') {
    return (
      <div
        aria-live="polite"
        className="flex w-full items-center gap-2 transition-opacity duration-200 motion-reduce:transition-none"
        style={{ opacity: visible ? 1 : 0 }}
      >
        <Icon variant="Loading" size={14} className="shrink-0" />
        <Text variant="body" weight="strong">
          {copy.title}
        </Text>
        {detail ? (
          <Text variant="subtext" theme="neutral" family="mono">
            {detail}
          </Text>
        ) : null}
      </div>
    )
  }

  return (
    <div
      aria-live="polite"
      className="flex w-full max-w-2xl flex-col items-center gap-5 text-center transition-opacity duration-200 motion-reduce:transition-none"
      style={{ opacity: visible ? 1 : 0 }}
    >
      <Icon variant="Loading" size={28} />
      <div className="flex flex-col items-center gap-1.5">
        {step >= 0 ? (
          <Text variant="subtext" theme="neutral">
            Step {step + 1} of {stages.length}
          </Text>
        ) : null}
        <Text variant="h3" weight="strong">
          {copy.title}
        </Text>
        {copy.body ? (
          <Text variant="body" theme="neutral">
            {copy.body}
          </Text>
        ) : null}
        {detail ? (
          <Text variant="subtext" theme="neutral" family="mono">
            {detail}
          </Text>
        ) : null}
      </div>
      {showBuilds ? (
        <div className="w-full overflow-hidden rounded-[10px] border text-left divide-y">
          {builds.map((build, index) => {
            const rowId = String(build.component_id || index)
            return (
              <BuildRow
                key={rowId}
                rowId={rowId}
                build={build}
                type={
                  build.component_type === 'sandbox'
                    ? undefined
                    : build.component_type
                }
                orgId={orgId}
                appId={appId}
                componentBuildId={build.build_id}
                sandboxBuildId={sandboxBuildId}
              />
            )
          })}
        </div>
      ) : null}
    </div>
  )
}
