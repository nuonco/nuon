import { Badge } from '@/components/common/Badge'
import type { TBadgeTheme } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TDiffNode } from '@/lib/ctl-api/apps/get-app-config-diff'

const SECTION_CONFIG: Record<string, { displayName: string; icon: TIconVariant; grouped: boolean }> = {
  components: { displayName: 'Components', icon: 'CubeIcon', grouped: true },
  actions: { displayName: 'Actions', icon: 'LightningIcon', grouped: true },
  inputs: { displayName: 'Install inputs', icon: 'ListBulletsIcon', grouped: true },
  secrets: { displayName: 'Secrets', icon: 'KeyIcon', grouped: true },
  sandbox: { displayName: 'Sandbox', icon: 'TerminalWindowIcon', grouped: false },
  runner: { displayName: 'Runner', icon: 'GearIcon', grouped: false },
  permissions: { displayName: 'Permissions', icon: 'ShieldIcon', grouped: false },
  stack: { displayName: 'Stack', icon: 'StackIcon', grouped: false },
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

export type DiffFieldEntry = {
  key: string
  op: string
  diff: string
}

export type DiffEntityEntry = {
  name: string
  op: 'add' | 'remove' | 'change'
  componentType?: string
  fields: DiffFieldEntry[]
}

export type DiffSectionData = {
  name: string
  sectionKey: string
  additions: number
  removals: number
  changed: number
  grouped: boolean
  entities: DiffEntityEntry[]
  fields: DiffFieldEntry[]
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

const DIFF_STYLES: Record<string, string> = {
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

function getEntityOp(node: TDiffNode): 'add' | 'remove' | 'change' {
  if (!node.children) {
    return (node.diff?.op as 'add' | 'remove' | 'change') || 'change'
  }

  let hasAdd = false
  let hasRemove = false
  let hasChange = false
  let allAdd = true
  let allRemove = true

  const walk = (n: TDiffNode) => {
    if (n.diff && n.diff.op !== 'noop' && n.diff.op !== '') {
      if (n.diff.op === 'add') hasAdd = true
      else { allAdd = false }
      if (n.diff.op === 'remove') hasRemove = true
      else { allRemove = false }
      if (n.diff.op === 'change') hasChange = true
    }
    if (n.children) n.children.forEach(walk)
  }
  node.children.forEach(walk)

  if (hasAdd && allAdd) return 'add'
  if (hasRemove && allRemove) return 'remove'
  return 'change'
}

function collectFields(node: TDiffNode): DiffFieldEntry[] {
  const fields: DiffFieldEntry[] = []

  const walk = (n: TDiffNode) => {
    if (n.diff && n.diff.op !== 'noop' && n.diff.op !== '') {
      fields.push({ key: n.key, op: n.diff.op, diff: n.diff.diff })
    }
    if (n.children) n.children.forEach(walk)
  }

  if (node.children) node.children.forEach(walk)
  return fields
}

function findComponentType(node: TDiffNode): string | undefined {
  if (!node.children) return undefined
  for (const child of node.children) {
    if (child.key === 'type' && child.diff) {
      const val = child.diff.diff
      const matches = [...val.matchAll(/'([^']+)'/g)]
      if (matches.length > 0) return matches[matches.length - 1][1]
    }
  }
  return undefined
}

export function extractSections(node?: TDiffNode): DiffSectionData[] {
  if (!node?.children) return []

  const sections: DiffSectionData[] = []
  for (const child of node.children) {
    const config = SECTION_CONFIG[child.key]
    if (!config) continue
    if (!child.children || child.children.length === 0) continue

    const section: DiffSectionData = {
      name: config.displayName,
      sectionKey: child.key,
      additions: 0,
      removals: 0,
      changed: 0,
      grouped: config.grouped,
      entities: [],
      fields: [],
    }

    if (config.grouped) {
      for (const entityNode of child.children) {
        const op = getEntityOp(entityNode)
        const fields = collectFields(entityNode)
        if (fields.length === 0) continue

        const componentType = child.key === 'components' ? findComponentType(entityNode) : undefined

        section.entities.push({
          name: entityNode.key,
          op,
          componentType,
          fields,
        })

        if (op === 'add') section.additions++
        else if (op === 'remove') section.removals++
        else section.changed++
      }
    } else {
      const fields = collectFields(child)
      section.fields = fields
      for (const f of fields) {
        if (f.op === 'add') section.additions++
        else if (f.op === 'remove') section.removals++
        else if (f.op === 'change') section.changed++
      }
    }

    if (section.entities.length > 0 || section.fields.length > 0) {
      sections.push(section)
    }
  }
  return sections
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
        <Text variant="base" theme="success" weight="strong">{summary.added}</Text>
        <Text variant="subtext" theme="neutral">to add</Text>
      </div>
      <div className="flex items-center gap-1.5">
        <Text variant="base" theme="warn" weight="strong">{summary.changed}</Text>
        <Text variant="subtext" theme="neutral">to change</Text>
      </div>
      <div className="flex items-center gap-1.5">
        <Text variant="base" theme="error" weight="strong">{summary.removed}</Text>
        <Text variant="subtext" theme="neutral">to remove</Text>
      </div>
    </div>
  </div>
)

const FieldsDiff = ({ fields }: { fields: DiffFieldEntry[] }) => (
  <div className="p-4 bg-code border-t shadow-xs min-h-[3rem] max-h-[40rem] overflow-auto font-mono text-[13px] leading-6">
    <div className="min-w-fit">
      {fields.map((field, idx) => {
        const prefix = getDiffPrefix(field.op)
        return (
          <div className={`flex whitespace-pre ${prefix.style}`} key={`${field.key}-${idx}`}>
            <span className="inline-block w-[2ch] shrink-0 select-none text-right mr-2 opacity-70">
              {prefix.char}
            </span>
            <span>
              <span className="font-semibold">{field.key}:</span>
              {'  '}
              <span>{field.diff}</span>
            </span>
          </div>
        )
      })}
    </div>
  </div>
)

const ComponentIcon = ({ type }: { type?: string }) => {
  if (!type) return null
  const config = COMPONENT_TYPE_ICON[type]
  if (!config) return null
  return <Icon variant={config.icon} size="14" className={config.brandClass} />
}

const EntityRow = ({ entity, sectionKey, idx }: { entity: DiffEntityEntry; sectionKey: string; idx: number }) => {
  const bgColor = getOpBgColor(entity.op)
  const borderColor = getOpBorderColor(entity.op)
  const isComponent = sectionKey === 'components'

  return (
    <Expand
      id={`${sectionKey}-${entity.name}-${idx}`}
      className={`border-l-4 ${borderColor}`}
      headerClassName={`w-full px-4 py-3 gap-3 text-left focus:outline-none ${bgColor}`}
      heading={
        <div className="text-left w-full">
          <div className="flex items-start justify-between w-full">
            <div className="flex items-center gap-2 max-w-[500px]">
              {isComponent && <ComponentIcon type={entity.componentType} />}
              <Text nowrap className="block truncate" weight="strong">
                {entity.name}
              </Text>
              {isComponent && entity.componentType && (
                <Text variant="subtext" theme="neutral">
                  {entity.componentType.replace(/_/g, ' ')}
                </Text>
              )}
            </div>
            <div className="flex items-center pr-4 self-center">
              <Badge theme={OP_BADGE_THEME[entity.op] || 'neutral'} size="sm">
                {entity.op}
              </Badge>
            </div>
          </div>
        </div>
      }
    >
      <FieldsDiff fields={entity.fields} />
    </Expand>
  )
}

const FieldRow = ({ field, sectionKey, idx }: { field: DiffFieldEntry; sectionKey: string; idx: number }) => {
  const bgColor = getOpBgColor(field.op)
  const borderColor = getOpBorderColor(field.op)

  return (
    <div className={`flex items-center justify-between border-l-4 px-4 py-3 ${borderColor} ${bgColor}`}>
      <div className="flex items-center gap-2">
        <Text weight="strong">{field.key}</Text>
        <Text variant="subtext" theme="neutral" family="mono">{field.diff}</Text>
      </div>
      <Badge theme={OP_BADGE_THEME[field.op as AppConfigOp] || 'neutral'} size="sm">
        {field.op}
      </Badge>
    </div>
  )
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
              {section.grouped
                ? section.entities.map((entity, idx) => (
                    <EntityRow key={`${entity.name}-${idx}`} entity={entity} sectionKey={section.sectionKey} idx={idx} />
                  ))
                : section.fields.map((field, idx) => (
                    <FieldRow key={`${field.key}-${idx}`} field={field} sectionKey={section.sectionKey} idx={idx} />
                  ))
              }
            </div>
          </Card>
        )
      })}
    </div>
  )
}
