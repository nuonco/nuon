import type { Node } from '@xyflow/react'
import type {
  TInstall,
  TInstallComponent,
  TInstallStack,
  TAppConfig,
  TComponentType,
} from '@/types'
import type { TInstallAppPermissionsConfig } from '@/lib/ctl-api/installs/get-install-app-permissions-config'
import { getStatusTheme } from '@/utils/status-utils'

export type TDiagramData = {
  install: TInstall
  components: TInstallComponent[]
  stack?: TInstallStack
  appConfig?: TAppConfig
  permissionsConfig?: TInstallAppPermissionsConfig
  orgId: string
}

export type TRoleInfo = {
  id: string
  name: string
  displayName: string
  description: string
  enabled: boolean
  policies: Array<{ name?: string; contents?: string }>
}

const CARD_W = 280
const CARD_H = 64
const CARD_GAP = 12
const PAD = 16
const HEADER = 48
const COLS = 2
const ROLE_W = 200
const ROLE_H = 48
const ROLE_GAP = 8

function gridDimensions(count: number) {
  if (count === 0) return { w: 0, h: 0 }
  const rows = Math.ceil(count / COLS)
  return {
    w: COLS * CARD_W + (COLS - 1) * CARD_GAP + PAD * 2,
    h: rows * CARD_H + (rows - 1) * CARD_GAP + PAD + HEADER,
  }
}

function cardXY(index: number, containerX: number, containerY: number) {
  const col = index % COLS
  const row = Math.floor(index / COLS)
  return {
    x: containerX + PAD + col * (CARD_W + CARD_GAP),
    y: containerY + HEADER + row * (CARD_H + CARD_GAP),
  }
}

export function extractRoles(permissionsConfig?: TInstallAppPermissionsConfig): TRoleInfo[] {
  if (!permissionsConfig) return []

  const entries = [
    { role: permissionsConfig.provision_role, fallback: 'Provision' },
    { role: permissionsConfig.deprovision_role, fallback: 'Deprovision' },
    { role: permissionsConfig.maintenance_role, fallback: 'Maintenance' },
    ...(permissionsConfig.break_glass_roles || []).map((r) => ({
      role: r,
      fallback: r?.display_name || r?.name || 'Break Glass',
    })),
    ...(permissionsConfig.custom_roles || []).map((r) => ({
      role: r,
      fallback: r?.display_name || r?.name || 'Custom',
    })),
  ]

  return entries
    .filter((e) => e.role)
    .map(({ role, fallback }) => ({
      id: role!.id || fallback,
      name: fallback,
      displayName: role!.display_name || fallback,
      description: role!.description || '',
      enabled: role!.enabled,
      policies: (role!.policies || []).map((p) => ({
        name: p.name,
        contents: p.contents,
      })),
    }))
}

const CLUSTER_TYPES: TComponentType[] = ['helm_chart', 'kubernetes_manifest']

function buildComponentNodeData(
  comp: TInstallComponent,
  isDrifted: boolean,
  orgId: string,
  installId: string
) {
  const latestDeploy = comp.install_deploys?.[0]
  return {
    name: comp.component?.name || comp.component_id || 'unknown',
    componentType: (comp.component?.type || 'unknown') as TComponentType,
    status: comp.status_v2?.status || '',
    statusDescription: comp.status_v2?.status_human_description || '',
    isDrifted,
    width: CARD_W,
    componentId: comp.component_id,
    installComponentId: comp.id,
    latestDeployStatus: latestDeploy?.status_v2?.status,
    latestDeployType: latestDeploy?.install_deploy_type,
    latestDeployAt: latestDeploy?.created_at,
    configVersion: comp.component?.config_versions,
    updatedAt: comp.updated_at,
    href: comp.component_id && installId ? `/${orgId}/installs/${installId}/components/${comp.component_id}` : undefined,
    deploysHref: comp.component_id && installId ? `/${orgId}/installs/${installId}/components/${comp.component_id}/deploys` : undefined,
  }
}

export function computeLayout(data: TDiagramData): Node[] {
  const { install, components, stack, permissionsConfig, orgId } = data
  const nodes: Node[] = []

  const roles = extractRoles(permissionsConfig)
  const clusterComps = components.filter((c) =>
    CLUSTER_TYPES.includes(c.component?.type as TComponentType)
  )
  const sandboxComps = components.filter(
    (c) => !CLUSTER_TYPES.includes(c.component?.type as TComponentType)
  )

  const driftSet = new Set(
    (install.drifted_objects || [])
      .map((d) => d.install_component_id)
      .filter(Boolean)
  )

  const eksGrid = gridDimensions(clusterComps.length)
  const sbGrid = gridDimensions(sandboxComps.length)

  const innerW = Math.max(eksGrid.w, sbGrid.w, 300)

  const sandboxInnerH =
    (eksGrid.h > 0 ? eksGrid.h + CARD_GAP : 0) +
    (sbGrid.h > 0 ? sbGrid.h : 0)
  const sandboxW = innerW + PAD * 2
  const sandboxH = Math.max(sandboxInnerH + HEADER + PAD, HEADER + PAD + 60)

  const vpcW = sandboxW + PAD * 2
  const vpcH = sandboxH + HEADER + PAD

  const hasRoles = roles.length > 0
  const rolesColW = hasRoles ? ROLE_W : 0
  const rolesGap = hasRoles ? PAD : 0
  const rolesBlockH =
    hasRoles
      ? roles.length * ROLE_H + (roles.length - 1) * ROLE_GAP
      : 0

  const contentW = rolesColW + rolesGap + vpcW
  const stackW = contentW + PAD * 2
  const contentH = Math.max(vpcH, rolesBlockH + 24)
  const stackH = contentH + HEADER + PAD

  const stackX = 0
  const stackY = 0
  const vpcX = stackX + PAD + rolesColW + rolesGap
  const vpcY = stackY + HEADER
  const sandboxX = vpcX + PAD
  const sandboxY = vpcY + HEADER
  const eksX = sandboxX + PAD
  const eksY = sandboxY + HEADER

  const latestVersion = stack?.versions?.[0]
  nodes.push({
    id: 'stack',
    type: 'containerNode',
    position: { x: stackX, y: stackY },
    data: {
      label: 'Stack',
      icon: 'StackSimpleIcon',
      status: latestVersion?.composite_status?.status || '',
      width: stackW,
      height: stackH,
      level: 0,
    },
    draggable: false,
    selectable: false,
  })

  if (roles.length > 0) {
    const rolesLabelX = stackX + PAD
    const rolesLabelY = stackY + HEADER

    nodes.push({
      id: 'iam-roles-label',
      type: 'sectionLabelNode',
      position: { x: rolesLabelX, y: rolesLabelY },
      data: { label: 'IAM Roles', icon: 'ShieldCheckIcon' },
      draggable: false,
      selectable: false,
    })

    for (let i = 0; i < roles.length; i++) {
      const role = roles[i]
      nodes.push({
        id: `role-${role.id}`,
        type: 'roleCardNode',
        position: {
          x: rolesLabelX,
          y: rolesLabelY + 24 + i * (ROLE_H + ROLE_GAP),
        },
        data: { ...role, width: ROLE_W },
        draggable: false,
        selectable: false,
      })
    }
  }

  nodes.push({
    id: 'vpc',
    type: 'containerNode',
    position: { x: vpcX, y: vpcY },
    data: {
      label: 'VPC',
      icon: 'CloudIcon',
      status: install.sandbox_status || '',
      width: vpcW,
      height: vpcH,
      level: 1,
    },
    draggable: false,
    selectable: false,
  })

  nodes.push({
    id: 'sandbox',
    type: 'containerNode',
    position: { x: sandboxX, y: sandboxY },
    data: {
      label: 'Sandbox',
      icon: 'CubeIcon',
      status: install.sandbox_status || '',
      width: sandboxW,
      height: sandboxH,
      level: 2,
      href: install.id ? `/${orgId}/installs/${install.id}/sandbox` : undefined,
    },
    draggable: false,
    selectable: false,
  })

  if (clusterComps.length > 0) {
    const clusterStatuses = clusterComps
      .map((c) => c.status_v2?.status || '')
      .filter(Boolean)
    const themePriority = ['error', 'warn', 'info', 'success', 'neutral'] as const
    const clusterStatus =
      themePriority.reduce<string>((found, theme) => {
        if (found) return found
        return clusterStatuses.find((s) => getStatusTheme(s) === theme) || ''
      }, '')

    nodes.push({
      id: 'eks-cluster',
      type: 'containerNode',
      position: { x: eksX, y: eksY },
      data: {
        label: 'EKS Cluster',
        icon: 'Kubernetes',
        status: clusterStatus,
        width: eksGrid.w,
        height: eksGrid.h,
        level: 3,
      },
      draggable: false,
      selectable: false,
    })

    for (let i = 0; i < clusterComps.length; i++) {
      const comp = clusterComps[i]
      const pos = cardXY(i, eksX, eksY)
      nodes.push({
        id: `comp-${comp.id || `cluster-${i}`}`,
        type: 'componentCardNode',
        position: pos,
        data: buildComponentNodeData(comp, driftSet.has(comp.id || ''), orgId, install.id || ''),
        draggable: false,
        selectable: false,
      })
    }
  }

  const sbCardsBaseY = eksY + (clusterComps.length > 0 ? eksGrid.h + CARD_GAP : 0)

  for (let i = 0; i < sandboxComps.length; i++) {
    const comp = sandboxComps[i]
    const col = i % COLS
    const row = Math.floor(i / COLS)
    nodes.push({
      id: `comp-${comp.id || i}`,
      type: 'componentCardNode',
      position: {
        x: sandboxX + PAD + col * (CARD_W + CARD_GAP),
        y: sbCardsBaseY + row * (CARD_H + CARD_GAP),
      },
      data: buildComponentNodeData(comp, driftSet.has(comp.id || ''), orgId, install.id || ''),
      draggable: false,
      selectable: false,
    })
  }

  return nodes
}
