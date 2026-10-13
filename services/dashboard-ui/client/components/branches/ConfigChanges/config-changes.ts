import type { TIconVariant } from '@/components/common/Icon'
import {
  emptyDiffSummary,
  MISSING_DIFF_ERROR,
  normalizeDiffOperation,
  type IPlanDiffSection,
  type IPlanDiffSummary,
  type TDiffOperation,
} from '@/lib/diffs'
import type {
  DiffEntityEntry,
  DiffFieldEntry,
  DiffFileEntry,
  DiffSectionData,
} from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'

export const CONFIG_CHANGE_OPERATIONS = ['create', 'update', 'delete'] as const

export type TConfigSourceFile = {
  path: string
  kind: string
  change: 'added' | 'modified' | 'removed' | 'unchanged'
  before?: string
  after?: string
}

export type TTemplateBuildChange = {
  id: string
  name: string
  kind: 'component' | 'sandbox'
  status: string
  changeReason?: string
  href?: string
  commit?: {
    sha: string
    message?: string
    author?: string
    repo?: string
  }
}

export type TConfigChangeSection = IPlanDiffSection & {
  group: string
  icon: TIconVariant
  kind: 'config' | 'file'
  build?: TTemplateBuildChange
}

export type TConfigChanges = {
  sections: TConfigChangeSection[]
  summary: IPlanDiffSummary
}

const SECTION_ICONS: Record<string, TIconVariant> = {
  components: 'CubeIcon',
  actions: 'LightningIcon',
  runbooks: 'BookOpenIcon',
  inputs: 'ListBulletsIcon',
  secrets: 'KeyIcon',
  sandbox: 'TerminalWindowIcon',
  runner: 'GearIcon',
  permissions: 'ShieldIcon',
  policies: 'ShieldCheckIcon',
  stack: 'StackIcon',
}

const SOURCE_FILES_GROUP = 'Source files'
const COMPONENTS_GROUP = 'Components'
const SANDBOX_GROUP = 'Sandbox'

const sourceOnlyConfig = (entity: DiffEntityEntry) => {
  if (entity.content) return false
  if (entity.changeKinds?.includes('config')) return false
  if (entity.changeKinds?.includes('source')) return true
  return (
    entity.fields.length > 0 &&
    entity.fields.every((field) => field.key === 'source')
  )
}

const awaitingBuild = (section: TConfigChangeSection) =>
  section.kind === 'config' &&
  !section.before &&
  !section.after &&
  !section.error

const withBuild = (
  section: TConfigChangeSection,
  build: TTemplateBuildChange
): TConfigChangeSection => ({
  ...section,
  build,
  searchable: [
    ...section.searchable,
    build.changeReason ?? '',
    build.commit?.sha ?? '',
    build.commit?.message ?? '',
    build.commit?.author ?? '',
  ],
})

const componentBuildSection = (
  build: TTemplateBuildChange
): TConfigChangeSection => ({
  id: `components/${build.name}`,
  kind: 'config',
  title: build.name,
  operation: 'update',
  before: '',
  after: '',
  language: 'toml',
  group: COMPONENTS_GROUP,
  icon: 'CubeIcon',
  build,
  searchable: [
    COMPONENTS_GROUP,
    build.name,
    build.changeReason ?? '',
    build.commit?.sha ?? '',
    build.commit?.message ?? '',
    build.commit?.author ?? '',
  ],
})

const sandboxBuildSection = (
  build: TTemplateBuildChange,
  id = 'sandbox'
): TConfigChangeSection => ({
  id,
  kind: 'config',
  title: build.name || SANDBOX_GROUP,
  operation: 'update',
  before: '',
  after: '',
  language: 'toml',
  group: SANDBOX_GROUP,
  icon: SECTION_ICONS.sandbox,
  build,
  searchable: [
    SANDBOX_GROUP,
    build.name,
    build.changeReason ?? '',
    build.commit?.sha ?? '',
    build.commit?.message ?? '',
    build.commit?.author ?? '',
  ],
})

const insertIntoGroup = (
  sections: TConfigChangeSection[],
  items: TConfigChangeSection[],
  group: string
) => {
  if (!items.length) return sections
  let last = -1
  for (let index = 0; index < sections.length; index++) {
    if (sections[index].group === group) last = index
  }
  if (last >= 0) {
    return [
      ...sections.slice(0, last + 1),
      ...items,
      ...sections.slice(last + 1),
    ]
  }
  if (group === COMPONENTS_GROUP) return [...items, ...sections]
  const source = sections.findIndex(
    (section) => section.group === SOURCE_FILES_GROUP
  )
  if (source < 0) return [...sections, ...items]
  return [...sections.slice(0, source), ...items, ...sections.slice(source)]
}

const attachBuilds = (
  sections: TConfigChangeSection[],
  builds: TTemplateBuildChange[]
): TConfigChangeSection[] => {
  const components = new Map<string, TTemplateBuildChange>()
  const sandboxes: TTemplateBuildChange[] = []
  for (const build of builds) {
    if (build.kind === 'sandbox') sandboxes.push(build)
    else if (!components.has(build.name)) components.set(build.name, build)
  }

  const attached: TConfigChangeSection[] = []
  for (const section of sections) {
    const component =
      section.id === `components/${section.title}`
        ? components.get(section.title)
        : undefined
    if (component) {
      components.delete(section.title)
      attached.push(withBuild(section, component))
      continue
    }
    const sandbox = section.id === 'sandbox' ? sandboxes.shift() : undefined
    if (sandbox) {
      attached.push(withBuild(section, sandbox))
      continue
    }
    if (awaitingBuild(section)) continue
    attached.push(section)
  }

  const sandboxAttached = attached.some((section) => section.id === 'sandbox')
  return insertIntoGroup(
    insertIntoGroup(
      attached,
      [...components.values()].map(componentBuildSection),
      COMPONENTS_GROUP
    ),
    sandboxes.map((build, index) =>
      sandboxBuildSection(
        build,
        !sandboxAttached && index === 0 ? 'sandbox' : `sandbox/${build.id}`
      )
    ),
    SANDBOX_GROUP
  )
}

const SOURCE_FILE_OPERATION: Record<
  Exclude<TConfigSourceFile['change'], 'unchanged'>,
  TDiffOperation
> = {
  added: 'create',
  modified: 'update',
  removed: 'delete',
}

const operationFor = (op: string): TDiffOperation =>
  normalizeDiffOperation(op) ?? 'update'

const languageFor = (path: string) => {
  const lower = path.toLowerCase()
  if (lower.endsWith('.toml')) return 'toml'
  if (lower.endsWith('.json')) return 'json'
  if (lower.endsWith('.tf') || lower.endsWith('.tfvars')) return 'hcl'
  if (lower.endsWith('.sh')) return 'shellscript'
  if (lower.endsWith('dockerfile')) return 'docker'
  return 'yaml'
}

const unquote = (value: string) => value.trim().replace(/^'([\s\S]*)'$/, '$1')

const tomlValue = (value: string) =>
  /^(true|false|-?\d+(\.\d+)?)$/.test(value) ? value : JSON.stringify(value)

const splitField = ({ diff, op }: DiffFieldEntry) => {
  const index = diff.indexOf(' -> ')
  if (index < 0) {
    return op === 'remove'
      ? { before: unquote(diff), after: '' }
      : { before: '', after: unquote(diff) }
  }
  return {
    before: unquote(diff.slice(0, index)),
    after: unquote(diff.slice(index + 4)),
  }
}

const tomlBlock = (
  name: string | undefined,
  fields: DiffFieldEntry[],
  side: 'before' | 'after'
) => {
  const lines = fields.flatMap((field) => {
    const value = splitField(field)[side]
    return value ? [`${field.key} = ${tomlValue(value)}`] : []
  })
  if (!lines.length) return ''
  return [name ? `name = ${tomlValue(name)}` : undefined, ...lines]
    .filter(Boolean)
    .join('\n')
}

const complete = (section: TConfigChangeSection): TConfigChangeSection =>
  section.before || section.after || section.error
    ? section
    : { ...section, error: MISSING_DIFF_ERROR }

const fileSection = (
  file: DiffFileEntry,
  id: string,
  group: string,
  icon: TIconVariant,
  description: string
): TConfigChangeSection =>
  complete({
    id,
    kind: 'file',
    title: file.name,
    description,
    operation: operationFor(file.op),
    before: file.before ?? '',
    after: file.after ?? '',
    language: languageFor(file.name),
    filename: file.name,
    group,
    icon,
    searchable: [group, description, file.name],
  })

const entitySections = (
  section: DiffSectionData,
  entity: DiffEntityEntry,
  icon: TIconVariant
): TConfigChangeSection[] => {
  const id = `${section.sectionKey}/${entity.name}`
  const files = (entity.files ?? []).map((file) =>
    fileSection(file, `${id}/${file.name}`, section.name, icon, entity.name)
  )
  if (sourceOnlyConfig(entity)) {
    return [
      {
        id,
        kind: 'config',
        title: entity.name,
        description: entity.componentType?.replace(/_/g, ' '),
        operation: 'update',
        before: '',
        after: '',
        language: 'toml',
        group: section.name,
        icon,
        searchable: [section.name, entity.name, entity.componentType ?? ''],
      },
      ...files,
    ]
  }
  const config = complete({
    id,
    kind: 'config',
    title: entity.name,
    description: entity.componentType?.replace(/_/g, ' '),
    operation: operationFor(entity.op),
    before:
      entity.content?.before ?? tomlBlock(entity.name, entity.fields, 'before'),
    after:
      entity.content?.after ?? tomlBlock(entity.name, entity.fields, 'after'),
    language: 'toml',
    filename: `${id}.toml`,
    group: section.name,
    icon,
    searchable: [
      section.name,
      entity.name,
      entity.componentType ?? '',
      ...entity.fields.flatMap(({ key, diff }) => [key, diff]),
    ],
  })
  return [config, ...files]
}

const sectionChanges = (section: DiffSectionData): TConfigChangeSection[] => {
  const icon = SECTION_ICONS[section.sectionKey] ?? 'CubeIcon'

  if (section.grouped) {
    return section.entities
      .filter(
        (entity) =>
          sourceOnlyConfig(entity) ||
          entity.fields.length ||
          entity.files?.length ||
          entity.content ||
          !entity.changeKinds ||
          entity.changeKinds.includes('config')
      )
      .flatMap((entity) => entitySections(section, entity, icon))
  }

  const op = section.content?.op ?? section.fields[0]?.op ?? 'change'
  const config = complete({
    id: section.sectionKey,
    kind: 'config',
    title: section.name,
    description: `${section.sectionKey}.toml`,
    operation: operationFor(op),
    before:
      section.content?.before ?? tomlBlock(undefined, section.fields, 'before'),
    after:
      section.content?.after ?? tomlBlock(undefined, section.fields, 'after'),
    language: 'toml',
    filename: `${section.sectionKey}.toml`,
    group: section.name,
    icon,
    searchable: [
      section.name,
      section.sectionKey,
      ...section.fields.flatMap(({ key, diff }) => [key, diff]),
    ],
  })
  const files = (section.files ?? []).map((file) =>
    fileSection(
      file,
      `${section.sectionKey}/${file.name}`,
      section.name,
      icon,
      section.name
    )
  )
  return [config, ...files]
}

const sourceFileSections = (
  files: TConfigSourceFile[]
): TConfigChangeSection[] =>
  files.flatMap((file) =>
    file.change === 'unchanged'
      ? []
      : [
          complete({
            id: `source/${file.path}`,
            kind: 'file',
            title: file.path,
            description: file.kind,
            operation: SOURCE_FILE_OPERATION[file.change],
            before: file.before ?? '',
            after: file.after ?? '',
            language: languageFor(file.path),
            filename: file.path,
            group: SOURCE_FILES_GROUP,
            icon: 'FileTextIcon',
            searchable: [SOURCE_FILES_GROUP, file.path, file.kind],
          }),
        ]
  )

export const configChanges = (
  sections: DiffSectionData[],
  files: TConfigSourceFile[] = [],
  builds: TTemplateBuildChange[] = []
): TConfigChanges => {
  const all = attachBuilds(
    [...sections.flatMap(sectionChanges), ...sourceFileSections(files)],
    builds
  )
  const summary = all
    .filter((section) => section.kind !== 'file')
    .reduce((counts, section) => {
      counts[section.operation] += 1
      return counts
    }, emptyDiffSummary())

  return { sections: all, summary }
}
