import { Badge } from '@/components/common/Badge'
import type { TBadgeTheme } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TDiffNode } from '@/lib/ctl-api/apps/get-app-config-diff'

const SECTION_CONFIG: Record<string, { displayName: string; icon: TIconVariant }> = {
  components: { displayName: 'Components', icon: 'CubeIcon' },
  actions: { displayName: 'Actions', icon: 'LightningIcon' },
  inputs: { displayName: 'Install inputs', icon: 'ListBulletsIcon' },
  secrets: { displayName: 'Secrets', icon: 'KeyIcon' },
  sandbox: { displayName: 'Sandbox', icon: 'TerminalWindowIcon' },
  runner: { displayName: 'Runner', icon: 'GearIcon' },
  permissions: { displayName: 'Permissions', icon: 'ShieldIcon' },
  stack: { displayName: 'Stack', icon: 'StackIcon' },
}

const COMPONENT_TYPE_ICON: Record<string, { icon: TIconVariant; brandClass: string }> = {
  helm_chart: { icon: 'Helm', brandClass: 'text-[#0F1689] dark:text-[#6A70D6]' },
  terraform_module: { icon: 'Terraform', brandClass: 'text-[#7B42BC] dark:text-[#A878E0]' },
  docker_build: { icon: 'Docker', brandClass: 'text-[#2496ED] dark:text-[#56B4F9]' },
  external_image: { icon: 'OCI', brandClass: 'text-[#262261] dark:text-[#8B87D1]' },
  kubernetes_manifest: { icon: 'Kubernetes', brandClass: 'text-[#326CE5] dark:text-[#5A8DEF]' },
  job: { icon: 'AWSLambda', brandClass: 'text-[#FF9900] dark:text-[#FFB340]' },
  pulumi: { icon: 'Pulumi', brandClass: 'text-[#8A3391] dark:text-[#B06AB8]' },
  pulumi_module: { icon: 'Pulumi', brandClass: 'text-[#8A3391] dark:text-[#C48BCC]' },
}

export type DiffSectionData = {
  name: string
  sectionKey: string
  additions: number
  removals: number
  changed: number
  entries: DiffEntry[]
}

export type DiffEntry = {
  op: string
  name: string
  description: string
}

type AppConfigOp = 'add' | 'remove' | 'change'

const OP_BADGE_THEME: Record<AppConfigOp, TBadgeTheme> = {
  add: 'success',
  remove: 'error',
  change: 'warn',
}

function getOpBgColor(op: string): string {
  switch (op) {
    case 'add':
      return [
        'bg-green-100 dark:bg-green-500/10',
        'hover:!bg-green-200 dark:hover:!bg-green-500/20',
        'focus:!bg-green-200 dark:focus:!bg-green-500/20',
        'active:!bg-green-300 dark:active:!bg-green-500/30',
      ].join(' ')
    case 'remove':
      return [
        'bg-red-100 dark:bg-red-500/10',
        'hover:!bg-red-200 dark:hover:!bg-red-500/20',
        'focus:!bg-red-200 dark:focus:!bg-red-500/20',
        'active:!bg-red-300 dark:active:!bg-red-500/30',
      ].join(' ')
    case 'change':
      return [
        'bg-orange-100 dark:bg-orange-500/10',
        'hover:!bg-orange-200 dark:hover:!bg-orange-500/20',
        'focus:!bg-orange-200 dark:focus:!bg-orange-500/20',
        'active:!bg-orange-300 dark:active:!bg-orange-500/30',
      ].join(' ')
    default:
      return [
        'bg-cool-grey-100 dark:bg-dark-grey-500/10',
        'hover:!bg-cool-grey-200 dark:hover:!bg-dark-grey-500/20',
        'focus:!bg-cool-grey-200 dark:focus:!bg-dark-grey-500/20',
        'active:!bg-cool-grey-300 dark:active:!bg-dark-grey-500/30',
      ].join(' ')
  }
}

function getOpBorderColor(op: string): string {
  switch (op) {
    case 'add':
      return '!border-l-green-400 dark:!border-l-green-600'
    case 'remove':
      return '!border-l-red-400 dark:!border-l-red-600'
    case 'change':
      return '!border-l-orange-400 dark:!border-l-orange-600'
    default:
      return '!border-l-cool-grey-400 dark:!border-l-cool-grey-500'
  }
}

const DIFF_STYLES = {
  add: 'bg-green-500/15 dark:bg-green-500/5 text-green-800 dark:text-green-400',
  remove: 'bg-red-500/15 dark:bg-red-500/5 text-red-800 dark:text-red-400',
  change: 'bg-orange-500/15 dark:bg-orange-500/5 text-orange-800 dark:text-orange-400',
}

function getDiffPrefix(op: string) {
  switch (op) {
    case 'add':
      return { char: '+', style: DIFF_STYLES.add }
    case 'remove':
      return { char: '-', style: DIFF_STYLES.remove }
    case 'change':
      return { char: '~', style: DIFF_STYLES.change }
    default:
      return { char: ' ', style: '' }
  }
}

export function extractSections(node?: TDiffNode): DiffSectionData[] {
  if (!node?.children) return []

  const sections: DiffSectionData[] = []
  for (const child of node.children) {
    const config = SECTION_CONFIG[child.key]
    if (!config) continue

    const section: DiffSectionData = {
      name: config.displayName,
      sectionKey: child.key,
      additions: 0,
      removals: 0,
      changed: 0,
      entries: [],
    }
    collectDiffEntries(child, '', section)
    if (section.entries.length > 0) {
      sections.push(section)
    }
  }
  return sections
}

function collectDiffEntries(node: TDiffNode, parentKey: string, section: DiffSectionData) {
  if (node.diff && node.diff.op !== 'noop' && node.diff.op !== '') {
    const entry: DiffEntry = {
      op: node.diff.op,
      name: parentKey || node.key,
      description: node.diff.diff,
    }
    if (node.diff.op === 'add') section.additions++
    else if (node.diff.op === 'remove') section.removals++
    else if (node.diff.op === 'change') section.changed++
    section.entries.push(entry)
    return
  }

  if (node.children) {
    const hasLeaves = node.children.some((c) => c.diff && c.diff.op !== 'noop' && c.diff.op !== '')
    if (hasLeaves) {
      for (const c of node.children) {
        if (c.diff && c.diff.op !== 'noop' && c.diff.op !== '') {
          const entry: DiffEntry = { op: c.diff.op, name: node.key, description: c.diff.diff }
          if (c.diff.op === 'add') section.additions++
          else if (c.diff.op === 'remove') section.removals++
          else if (c.diff.op === 'change') section.changed++
          section.entries.push(entry)
        }
      }
    } else {
      for (const c of node.children) {
        collectDiffEntries(c, node.key || parentKey, section)
      }
    }
  }
}

export function computeSummary(sections: DiffSectionData[]) {
  let added = 0, removed = 0, changed = 0
  for (const s of sections) {
    added += s.additions
    removed += s.removals
    changed += s.changed
  }
  return { added, removed, changed }
}

const AppConfigSummary = ({ summary }: { summary: { added: number; removed: number; changed: number } }) => (
  <div className="px-4 py-3 sm:px-6 border-b bg-cool-grey-100 dark:bg-dark-grey-800">
    <div className="flex space-x-4">
      <div className="flex items-center gap-1.5">
        <Text variant="base" theme="success" weight="strong">
          {summary.added}
        </Text>
        <Text variant="subtext" theme="neutral">
          to add
        </Text>
      </div>
      <div className="flex items-center gap-1.5">
        <Text variant="base" theme="warn" weight="strong">
          {summary.changed}
        </Text>
        <Text variant="subtext" theme="neutral">
          to change
        </Text>
      </div>
      <div className="flex items-center gap-1.5">
        <Text variant="base" theme="error" weight="strong">
          {summary.removed}
        </Text>
        <Text variant="subtext" theme="neutral">
          to remove
        </Text>
      </div>
    </div>
  </div>
)

const EntryValuesDiff = ({ entries }: { entries: DiffEntry[] }) => (
  <div className="p-4 bg-code border-t shadow-xs min-h-[3rem] max-h-[40rem] overflow-auto font-mono text-[13px] leading-6">
    <div className="min-w-fit">
      {entries.map((entry, idx) => {
        const prefix = getDiffPrefix(entry.op)
        return (
          <div className={`flex whitespace-pre ${prefix.style}`} key={`${entry.name}-${idx}`}>
            <span className="inline-block w-[2ch] shrink-0 select-none text-right mr-2 opacity-70">
              {prefix.char}
            </span>
            <span>
              <span className="font-semibold">{entry.name}:</span>
              {'  '}
              <span>{entry.description}</span>
            </span>
          </div>
        )
      })}
    </div>
  </div>
)

const EntryComponentIcon = ({ description }: { description: string }) => {
  const componentType = description.trim().toLowerCase().replace(/\s+/g, '_')
  const config = COMPONENT_TYPE_ICON[componentType]
  if (!config) return null
  return <Icon variant={config.icon} size="14" className={config.brandClass} />
}

export interface IAppConfigDiff {
  sections: DiffSectionData[]
  summary: { added: number; removed: number; changed: number } | null
  isLoading?: boolean
}

export const AppConfigDiff = ({
  sections,
  summary,
  isLoading = false,
}: IAppConfigDiff) => {
  if (isLoading) {
    return (
      <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
        <div className="px-4 sm:px-6 py-4">
          <Text variant="subtext" theme="neutral">Loading config diff...</Text>
        </div>
      </Card>
    )
  }

  if (sections.length === 0) {
    return (
      <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
        <div className="px-4 sm:px-6 py-4">
          <Text variant="subtext" theme="neutral">No config changes detected</Text>
        </div>
      </Card>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      {sections.map((section) => {
        const sectionIcon = SECTION_CONFIG[section.sectionKey]?.icon

        return (
          <Card key={section.name} className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
            <div className="px-4 sm:px-6 py-4 border-b">
              <Text flex className="gap-2 items-center" variant="base" weight="strong">
                {sectionIcon && <Icon variant={sectionIcon} size="16" />}
                {section.name}
              </Text>
            </div>

            <AppConfigSummary summary={{ added: section.additions, removed: section.removals, changed: section.changed }} />

            <div className="flex flex-col divide-y">
              {section.entries.map((entry, idx) => {
                const bgColor = getOpBgColor(entry.op)
                const borderColor = getOpBorderColor(entry.op)
                const isComponentSection = section.sectionKey === 'components'

                return (
                  <Expand
                    key={`${entry.name}-${idx}`}
                    id={`${section.name}-${entry.name}-${idx}`}
                    className={`border-l-4 ${borderColor}`}
                    headerClassName={`w-full px-4 py-3 gap-3 text-left focus:outline-none ${bgColor}`}
                    heading={
                      <div className="text-left w-full">
                        <div className="flex items-start justify-between w-full">
                          <div className="flex items-center gap-2 max-w-[500px]">
                            {isComponentSection && <EntryComponentIcon description={entry.description} />}
                            <Text nowrap className="block truncate" weight="strong">
                              {entry.name}
                            </Text>
                            {isComponentSection && (
                              <Text variant="subtext" theme="neutral">
                                {entry.description}
                              </Text>
                            )}
                          </div>
                          <div className="flex items-center pr-4 self-center">
                            <Badge theme={OP_BADGE_THEME[entry.op as AppConfigOp] || 'neutral'} size="sm">
                              {entry.op}
                            </Badge>
                          </div>
                        </div>
                      </div>
                    }
                  >
                    <EntryValuesDiff entries={[entry]} />
                  </Expand>
                )
              })}
            </div>
          </Card>
        )
      })}

      {summary && (
        <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
          <div className="px-4 sm:px-6 py-4 border-b">
            <Text variant="base" weight="strong">
              Total changes
            </Text>
          </div>
          <AppConfigSummary summary={summary} />
        </Card>
      )}
    </div>
  )
}
