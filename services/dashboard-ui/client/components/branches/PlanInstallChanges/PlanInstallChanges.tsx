import { useState } from 'react'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Menu } from '@/components/common/Menu'
import { SearchInput } from '@/components/common/SearchInput'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import {
  ConfigChangesViewer,
  configChanges,
  type TTemplateBuildChange,
} from '@/components/branches/ConfigChanges'
import { ConfigChangesLoading } from '@/components/branches/ConfigChanges/ConfigChangesLoading'
import type { PlanInstallDiff } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep/PlanGroupStep'
import { cn } from '@/utils/classnames'

export type TPlanInstallFacts = Record<
  string,
  {
    labels?: Record<string, string>
    region?: string
    status?: string
    detail?: string
    appliedConfigId?: string
  }
>

export interface IPlanInstallChanges {
  installs: PlanInstallDiff[]
  builds?: TTemplateBuildChange[]
  installFacts?: TPlanInstallFacts
  labelColors?: Record<string, string>
  isLoading?: boolean
}

const NO_BUILDS: TTemplateBuildChange[] = []

const buildsFor = (install: PlanInstallDiff, builds: TTemplateBuildChange[]) =>
  install.alreadyApplied ? NO_BUILDS : builds

const InstallCounts = ({
  install,
  builds,
}: {
  install: PlanInstallDiff
  builds: TTemplateBuildChange[]
}) => {
  if (install.isLoading) {
    return (
      <Text variant="label" theme="neutral">
        Loading
      </Text>
    )
  }
  const { summary } = configChanges(
    install.sections,
    [],
    buildsFor(install, builds)
  )
  return (
    <ChangeCountSummary
      added={summary.create}
      updated={summary.update}
      removed={summary.delete}
      emptyText="No changes"
    />
  )
}

export const PlanInstallChanges = ({
  installs,
  builds = NO_BUILDS,
  installFacts,
  labelColors,
  isLoading,
}: IPlanInstallChanges) => {
  const [selectedId, setSelectedId] = useState<string>()
  const [query, setQuery] = useState('')
  const selected =
    installs.find((install) => install.installId === selectedId) ?? installs[0]
  const search = query.trim().toLowerCase()
  const matches = search
    ? installs.filter((install) =>
        install.installName.toLowerCase().includes(search)
      )
    : installs

  if (isLoading) return <ConfigChangesLoading />

  if (!selected) {
    return (
      <EmptyState
        emptyTitle="No installs"
        emptyMessage="This install group plan has no installs."
        variant="diagram"
        size="sm"
      />
    )
  }

  const facts = installFacts?.[selected.installId]
  const labels = Object.entries(selected.installLabels ?? facts?.labels ?? {})

  return (
    <div className="flex flex-col gap-6 pt-1 md:min-h-0 md:flex-1">
      <div className="flex shrink-0 flex-wrap items-center gap-3">
        {installs.length > 1 ? (
          <Dropdown
            id="plan-install-changes"
            buttonText={
              <span className="flex items-center gap-2">
                <Icon variant="CubeIcon" size={14} />
                {selected.installName}
              </span>
            }
          >
            <Menu className="!w-80">
              <div onClick={(event) => event.stopPropagation()}>
                <SearchInput
                  labelClassName="w-full"
                  className="w-full md:min-w-0"
                  placeholder="Search installs"
                  value={query}
                  onChange={setQuery}
                />
              </div>
              <div className="-mx-2 -mb-2 flex max-h-80 flex-col gap-0.5 overflow-y-auto overscroll-y-contain p-2">
                {matches.length === 0 ? (
                  <Text
                    variant="subtext"
                    theme="neutral"
                    className="px-2 py-1.5"
                  >
                    No matching installs
                  </Text>
                ) : null}
                {matches.map((install) => (
                  <Button
                    key={install.installId}
                    type="button"
                    variant="ghost"
                    isMenuButton
                    className={cn(
                      'justify-between gap-4',
                      install.installId === selected.installId &&
                        'bg-cool-grey-500/12'
                    )}
                    onClick={() => setSelectedId(install.installId)}
                  >
                    <span className="flex min-w-0 flex-col items-start gap-0.5">
                      <span className="max-w-full truncate">
                        {install.installName}
                      </span>
                      {install.versionLabel ? (
                        <Text variant="label" theme="neutral" family="mono">
                          {install.versionLabel}
                        </Text>
                      ) : null}
                    </span>
                    <InstallCounts install={install} builds={builds} />
                  </Button>
                ))}
              </div>
            </Menu>
          </Dropdown>
        ) : (
          <Text variant="base" weight="strong">
            {selected.installName}
          </Text>
        )}
        {labels.map(([key, value]) => (
          <LabelBadge
            key={key}
            labelKey={key}
            labelValue={value}
            size="sm"
            customColor={labelColors?.[key]}
          />
        ))}
        {facts?.region ? (
          <Text variant="subtext" theme="neutral" family="mono">
            {facts.region}
          </Text>
        ) : null}
        {facts?.status ? <Status status={facts.status} /> : null}
      </div>
      {selected.isLoading ? (
        <ConfigChangesLoading />
      ) : selected.isError ? (
        <Text variant="subtext" theme="neutral">
          Changes could not be loaded.
        </Text>
      ) : (
        <ConfigChangesViewer
          key={selected.installId}
          sections={selected.sections}
          builds={buildsFor(selected, builds)}
          versionLabel={selected.versionLabel}
          previousSha={selected.previousSha}
          sha={selected.sha}
          message={selected.message}
          author={selected.author}
          createdAt={selected.createdAt}
          emptyMessage="This install already runs this config."
        />
      )}
    </div>
  )
}
