import type {
  TAppConfigDiffField,
  TAppConfigDiffOperation,
  TAppConfigDiffSection,
} from '@/types'
import {
  MISSING_DIFF_ERROR,
  emptyDiffSummary,
  normalizeDiffOperation,
  type IPlanDiffGroup,
  type IPlanDiffSection,
  type IPlanDiffSummary,
  type TDiffOperation,
} from '.'

export const APP_CONFIG_DIFF_OPERATIONS = [
  'create',
  'update',
  'delete',
] as const

export interface IAppConfigDiffSummary {
  added: number
  removed: number
  changed: number
}

const operationFor = (
  operation: TAppConfigDiffOperation | string
): TDiffOperation => normalizeDiffOperation(operation) ?? 'update'

const languageFor = (name: string) => {
  const lower = name.toLowerCase()
  if (lower.endsWith('.yaml') || lower.endsWith('.yml')) return 'yaml'
  if (lower.endsWith('.json')) return 'json'
  if (
    lower.endsWith('.tf') ||
    lower.endsWith('.tfvars') ||
    lower.startsWith('var_file')
  ) {
    return 'hcl'
  }
  if (lower === 'dockerfile') return 'docker'
  if (lower.endsWith('.sh') || lower.endsWith('inline_contents')) {
    return 'shellscript'
  }
  return 'yaml'
}

const splitFieldDiff = (field: TAppConfigDiffField) => {
  const separator = ' -> '
  const index = field.diff.indexOf(separator)
  if (index < 0) {
    return field.op === 'remove'
      ? { before: field.diff, after: '' }
      : { before: '', after: field.diff }
  }

  return {
    before: field.diff.slice(0, index),
    after: field.diff.slice(index + separator.length),
  }
}

const fieldContents = (fields: TAppConfigDiffField[]) => {
  const before: string[] = []
  const after: string[] = []

  fields.forEach((field) => {
    const values = splitFieldDiff(field)
    if (field.op !== 'add' && values.before) {
      before.push(`${field.key} = ${values.before}`)
    }
    if (field.op !== 'remove' && values.after) {
      after.push(`${field.key} = ${values.after}`)
    }
  })

  return { before: before.join('\n'), after: after.join('\n') }
}

const complete = (section: IPlanDiffSection): IPlanDiffSection =>
  section.before || section.after || section.error
    ? section
    : { ...section, error: MISSING_DIFF_ERROR }

const contentSection = (
  section: TAppConfigDiffSection,
  index: number
): IPlanDiffSection | undefined => {
  if (!section.content) return undefined

  return complete({
    id: `${section.sectionKey}/content/${index}`,
    title: section.name,
    description: section.sectionKey,
    operation: operationFor(section.content.op),
    before: section.content.before ?? '',
    after: section.content.after ?? '',
    language: 'toml',
    filename: `${section.sectionKey}.toml`,
    searchable: [section.name, section.sectionKey],
    group: section.name,
  })
}

const entitySections = (
  section: TAppConfigDiffSection,
  sectionIndex: number
): IPlanDiffSection[] =>
  section.entities.flatMap((entity, entityIndex) => {
    const identity = [
      section.name,
      entity.name,
      entity.componentType ?? '',
      section.sectionKey,
    ]
    const fields = entity.fields.length
      ? [
          complete({
            id: `${section.sectionKey}/${entity.name}/fields/${sectionIndex}/${entityIndex}`,
            title: entity.name,
            description: [section.name, entity.componentType]
              .filter(Boolean)
              .join(' · '),
            operation: operationFor(entity.op),
            ...fieldContents(entity.fields),
            language: 'toml',
            filename: `${entity.name}.toml`,
            group: section.name,
            searchable: [
              ...identity,
              ...entity.fields.flatMap(({ key, diff }) => [key, diff]),
            ],
          }),
        ]
      : []
    const files = (entity.files ?? []).map((file, fileIndex) =>
      complete({
        id: `${section.sectionKey}/${entity.name}/file/${file.name}/${sectionIndex}/${entityIndex}/${fileIndex}`,
        title: file.name,
        description: `${section.name} · ${entity.name}`,
        operation: operationFor(file.op),
        before: file.before ?? '',
        after: file.after ?? '',
        language: languageFor(file.name),
        filename: file.name,
        group: section.name,
        searchable: [...identity, file.name],
      })
    )

    return [...fields, ...files]
  })

const ungroupedSections = (
  section: TAppConfigDiffSection,
  sectionIndex: number
): IPlanDiffSection[] => {
  const fields = section.fields.map((field, fieldIndex) =>
    complete({
      id: `${section.sectionKey}/field/${field.key}/${sectionIndex}/${fieldIndex}`,
      title: field.key,
      description: section.name,
      operation: operationFor(field.op),
      ...fieldContents([field]),
      language: 'toml',
      filename: `${section.sectionKey}.toml`,
      group: section.name,
      searchable: [section.name, section.sectionKey, field.key, field.diff],
    })
  )
  const files = (section.files ?? []).map((file, fileIndex) =>
    complete({
      id: `${section.sectionKey}/file/${file.name}/${sectionIndex}/${fileIndex}`,
      title: file.name,
      description: section.name,
      operation: operationFor(file.op),
      before: file.before ?? '',
      after: file.after ?? '',
      language: languageFor(file.name),
      filename: file.name,
      group: section.name,
      searchable: [section.name, section.sectionKey, file.name],
    })
  )

  return [...fields, ...files]
}

const normalizedSummary = (
  summary: IAppConfigDiffSummary | null | undefined,
  sections: IPlanDiffSection[]
): IPlanDiffSummary => {
  if (!summary) {
    return sections.reduce((counts, section) => {
      counts[section.operation] += 1
      return counts
    }, emptyDiffSummary())
  }

  return {
    ...emptyDiffSummary(),
    create: summary.added,
    update: summary.changed,
    delete: summary.removed,
  }
}

export const appConfigPlanDiff = (
  source: TAppConfigDiffSection[],
  summary?: IAppConfigDiffSummary | null
): IPlanDiffGroup => {
  const sections = source.flatMap((section, index) => {
    const content = contentSection(section, index)
    if (content) return [content]
    return section.grouped
      ? entitySections(section, index)
      : ungroupedSections(section, index)
  })

  return {
    id: 'app-config',
    title: 'App config changes',
    searchPlaceholder: 'Search components, actions, fields, or files',
    sections,
    summary: normalizedSummary(summary, sections),
  }
}
