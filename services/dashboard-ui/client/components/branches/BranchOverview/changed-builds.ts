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
  changeReason?: string
}

const isSandbox = (build: TBuildMeta) =>
  build.component_type === 'sandbox' || build.component_id === 'sandbox'

export const changeReasonFor = (build: TBuildMeta) =>
  build.change_reason ||
  (build.skipped || build.status === 'skipped' ? 'no_changes' : 'source_changed')

export const buildChanged = (build: TBuildMeta) => {
  const reason = changeReasonFor(build)
  return (
    reason === 'source_changed' ||
    reason === 'config_changed' ||
    reason === 'source_and_config'
  )
}

const componentHref = (orgId: string, appId: string, componentId: string, buildId: string) =>
  `/${orgId}/apps/${appId}/components/${componentId}/builds/${buildId}`

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

  return metaBuilds.flatMap((build, index) => {
    if (!buildChanged(build)) return []
    if (isSandbox(build)) {
      return [
        {
          id: sandboxBuildId || 'sandbox',
          name: build.component_name || 'Sandbox',
          status: build.status || 'unknown',
          changeReason: changeReasonFor(build),
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
        href:
          build.component_id && buildId
            ? componentHref(orgId, appId, build.component_id, buildId)
            : undefined,
      },
    ]
  })
}
