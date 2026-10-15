export interface TBuildMeta {
  component_id?: string
  component_name?: string
  component_type?: string
  status?: string
  skipped?: boolean
  change_reason?: string
}

export interface TRunBuildRef {
  id?: string
  component_id?: string
  component_name?: string
  status?: string
  change_reason?: string
}

export interface TChangedBuildRow {
  id: string
  name: string
  status: string
  href?: string
  componentId?: string
  changeReason?: string
  kind?: 'component' | 'sandbox'
}

const isSandbox = (build: TBuildMeta) =>
  build.component_type === 'sandbox' || build.component_id === 'sandbox'

export const changeReasonFor = (build: TBuildMeta) =>
  build.change_reason ||
  (build.skipped || build.status === 'skipped'
    ? 'no_changes'
    : 'source_changed')

export const buildChanged = (build: TBuildMeta) => {
  const reason = changeReasonFor(build)
  return isConfigChange(reason) || isSourceChange(reason)
}

export const isConfigChange = (reason?: string) =>
  reason === 'config_changed' || reason === 'source_and_config'

export const isSourceChange = (reason?: string) =>
  reason === 'source_changed' || reason === 'source_and_config'

export const splitChangedBuilds = (rows: TChangedBuildRow[]) => ({
  config: rows.filter((row) => isConfigChange(row.changeReason)),
  source: rows.filter((row) => isSourceChange(row.changeReason)),
})

const componentHref = (
  orgId: string,
  appId: string,
  componentId: string,
  buildId: string
) => `/${orgId}/apps/${appId}/components/${componentId}/builds/${buildId}`

export interface TDiffComponentChange {
  name: string
  changeKinds?: Array<'source' | 'config'>
  fields?: { key: string }[]
}

const changeReasonFromDiff = (component: TDiffComponentChange) => {
  const kinds = component.changeKinds ?? []
  const sourceFromFields =
    kinds.length === 0 &&
    !!component.fields?.length &&
    component.fields.every((field) => field.key === 'source')
  const source = kinds.includes('source') || sourceFromFields
  const config = kinds.includes('config') || (kinds.length === 0 && !source)
  if (source && config) return 'source_and_config'
  if (source) return 'source_changed'
  return 'config_changed'
}

const matchesComponentName = (
  build: { component_name?: string; component_id?: string },
  name: string
) =>
  build.component_name === name ||
  (!build.component_name && build.component_id === name)

export const changedBuildRows = ({
  metaBuilds,
  runBuilds,
  orgId,
  appId,
  sandboxBuildId,
}: {
  metaBuilds: TBuildMeta[]
  runBuilds: TRunBuildRef[]
  orgId?: string
  appId?: string
  sandboxBuildId?: string
}): TChangedBuildRow[] => {
  if (!orgId || !appId) return []

  if (metaBuilds.length === 0) {
    return runBuilds.flatMap((build) => {
      if (!build.id) return []
      return [
        {
          id: build.id,
          name: build.component_name || build.component_id || 'Component',
          status: build.status || 'unknown',
          changeReason: build.change_reason,
          kind: 'component' as const,
          componentId: build.component_id,
          href: build.component_id
            ? componentHref(orgId, appId, build.component_id, build.id)
            : undefined,
        },
      ]
    })
  }

  const buildIdByComponent = new Map(
    runBuilds.flatMap((build) =>
      build.component_id && build.id
        ? [[build.component_id, build.id] as const]
        : []
    )
  )

  return metaBuilds.flatMap<TChangedBuildRow>((build, index) => {
    if (!buildChanged(build)) return []
    if (isSandbox(build)) {
      return [
        {
          id: sandboxBuildId || 'sandbox',
          name: build.component_name || 'Sandbox',
          status: build.status || 'unknown',
          changeReason: changeReasonFor(build),
          kind: 'sandbox' as const,
          href: sandboxBuildId
            ? `/${orgId}/apps/${appId}/sandbox/builds/${sandboxBuildId}`
            : undefined,
        },
      ]
    }
    const buildId = build.component_id
      ? buildIdByComponent.get(build.component_id)
      : undefined
    return [
      {
        id: buildId || build.component_id || `build-${index}`,
        name: build.component_name || build.component_id || 'Component',
        status: build.status || 'unknown',
        changeReason: changeReasonFor(build),
        kind: 'component' as const,
        componentId: build.component_id,
        href:
          build.component_id && buildId
            ? componentHref(orgId, appId, build.component_id, buildId)
            : undefined,
      },
    ]
  })
}

export const mergedChangedBuildRows = ({
  components = [],
  ...args
}: {
  metaBuilds: TBuildMeta[]
  runBuilds: TRunBuildRef[]
  orgId?: string
  appId?: string
  sandboxBuildId?: string
  components?: TDiffComponentChange[]
}): TChangedBuildRow[] => {
  const fromBuilds = changedBuildRows(args)
  if (components.length === 0) return fromBuilds
  if (!args.orgId || !args.appId) return []

  const seen = new Set<string>()
  const fromDiff: TChangedBuildRow[] = []
  const { orgId, appId, metaBuilds, runBuilds } = args

  for (const component of components) {
    const name = component.name
    if (!name || seen.has(name)) continue

    const meta = metaBuilds.find(
      (build) => !isSandbox(build) && matchesComponentName(build, name)
    )
    const run = runBuilds.find((build) => matchesComponentName(build, name))

    seen.add(name)
    if (meta?.component_name) seen.add(meta.component_name)
    if (meta?.component_id) seen.add(meta.component_id)
    if (run?.component_name) seen.add(run.component_name)
    if (run?.component_id) seen.add(run.component_id)

    const buildId = run?.id
    const componentId = buildId
      ? run?.component_id || meta?.component_id
      : undefined

    fromDiff.push({
      id: buildId || name,
      name,
      status: meta?.status || run?.status || 'pending',
      changeReason:
        meta?.change_reason ||
        run?.change_reason ||
        changeReasonFromDiff(component),
      kind: 'component',
      componentId,
      href:
        componentId && buildId
          ? componentHref(orgId, appId, componentId, buildId)
          : undefined,
    })
  }

  const rest = fromBuilds.filter(
    (row) =>
      row.kind === 'sandbox' ||
      (!seen.has(row.name) && !(row.componentId && seen.has(row.componentId)))
  )

  return [...fromDiff, ...rest]
}
