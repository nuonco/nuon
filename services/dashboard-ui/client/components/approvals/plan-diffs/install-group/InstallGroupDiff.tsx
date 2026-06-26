import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { AppConfigDiff } from '../app-config/AppConfigDiff'
import type { DiffSectionData } from '../app-config/AppConfigDiff'

export type InstallDiffEntry = {
  installId: string
  installName: string
  installLabels?: Record<string, string>
  status?: string
  sections: DiffSectionData[]
  summary: { added: number; removed: number; changed: number }
  sandboxChanged?: boolean
  stackChanged?: boolean
}

export interface IInstallGroupDiff {
  groupName: string
  installs: InstallDiffEntry[]
}

const InstallStatusDot = ({ status }: { status?: string }) => {
  if (status === 'success' || status === 'deployed') {
    return <div className="w-[10px] h-[10px] rounded-full bg-green-500 shrink-0" />
  }
  if (status === 'in-progress') {
    return <div className="w-[10px] h-[10px] rounded-full bg-blue-500 shrink-0 animate-pulse" />
  }
  if (status === 'error') {
    return <div className="w-[10px] h-[10px] rounded-full bg-red-500 shrink-0" />
  }
  return <div className="w-[10px] h-[10px] rounded-full bg-cool-grey-300 dark:bg-dark-grey-500 shrink-0" />
}

export const InstallGroupDiff = ({ groupName, installs }: IInstallGroupDiff) => {
  if (installs.length === 0) {
    return (
      <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
        <div className="px-4 sm:px-6 py-4">
          <Text variant="subtext" theme="neutral">No install changes to show</Text>
        </div>
      </Card>
    )
  }

  return (
    <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
      <div className="px-4 sm:px-6 py-4 border-b">
        <div className="flex items-center gap-3">
          <Icon variant="ListChecksIcon" size="16" />
          <Text variant="base" weight="strong">{groupName}</Text>
          <Text variant="subtext" theme="neutral">
            {installs.length} {installs.length === 1 ? 'install' : 'installs'}
          </Text>
        </div>
      </div>

      <div className="flex flex-col divide-y">
        {installs.map((install) => {
          const totalChanges = install.summary.added + install.summary.removed + install.summary.changed
          const hasChanges = totalChanges > 0 || install.sandboxChanged || install.stackChanged
          const labelEntries = install.installLabels ? Object.entries(install.installLabels) : []

          const heading = (
            <div className="flex items-center gap-3 w-full">
              <InstallStatusDot status={install.status} />
              <Text weight="strong">{install.installName || install.installId}</Text>
              {labelEntries.map(([k, v]) => (
                <span key={k} className="inline-flex items-center px-1.5 py-0.5 rounded border border-cool-grey-200 dark:border-dark-grey-600 bg-cool-grey-50 dark:bg-dark-grey-800 font-mono text-[10.5px] text-cool-grey-500 dark:text-cool-grey-400 shrink-0">
                  {k}={v}
                </span>
              ))}
              <div className="flex items-center gap-1.5 ml-auto shrink-0">
                {install.summary.added > 0 && (
                  <span className="text-[12px] font-semibold text-green-600 dark:text-green-400">+{install.summary.added}</span>
                )}
                {install.summary.changed > 0 && (
                  <span className="text-[12px] font-semibold text-yellow-600 dark:text-yellow-400">~{install.summary.changed}</span>
                )}
                {install.summary.removed > 0 && (
                  <span className="text-[12px] font-semibold text-red-600 dark:text-red-400">-{install.summary.removed}</span>
                )}
                {install.sandboxChanged && <Badge theme="warn" size="sm">sandbox</Badge>}
                {install.stackChanged && <Badge theme="warn" size="sm">stack</Badge>}
                {!hasChanges && (
                  <span className="text-[12px] text-cool-grey-400 dark:text-cool-grey-500">no changes</span>
                )}
              </div>
            </div>
          )

          if (!hasChanges || install.sections.length === 0) {
            return (
              <div key={install.installId} className="px-4 py-3">
                {heading}
              </div>
            )
          }

          return (
            <Expand
              key={install.installId}
              id={`install-group-diff-${install.installId}`}
              heading={heading}
              isIconBeforeHeading
              headerClassName="px-4 py-3"
            >
              <div className="px-4 pb-4">
                <AppConfigDiff
                  sections={install.sections}
                  summary={install.summary}
                />
              </div>
            </Expand>
          )
        })}
      </div>
    </Card>
  )
}
