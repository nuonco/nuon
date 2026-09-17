export const appSetupHref = (orgId: string, appId?: string) => {
  const path = `/${orgId}/apps/setup`
  if (!appId) return path
  return `${path}?${new URLSearchParams({ appId })}`
}

export const installSetupHref = (orgId: string, installId?: string) => {
  const path = `/${orgId}/installs/setup`
  if (!installId) return path
  return `${path}?${new URLSearchParams({ installId })}`
}

export const installHref = (orgId: string, installId: string) =>
  `/${orgId}/installs/${installId}`

export const appBranchHref = (orgId: string, appId: string, branchId: string) =>
  `/${orgId}/apps/${appId}/branches/${branchId}`

export const resolveAppHref = ({
  orgId,
  appId,
  branchIds,
  lastBranchId,
}: {
  orgId: string
  appId: string
  branchIds: string[]
  lastBranchId?: string
}) => {
  if (branchIds.length === 0) return appSetupHref(orgId, appId)
  if (branchIds.length === 1) {
    return appBranchHref(orgId, appId, branchIds[0]!)
  }
  if (lastBranchId && branchIds.includes(lastBranchId)) {
    return appBranchHref(orgId, appId, lastBranchId)
  }
  return appBranchHref(orgId, appId, branchIds[0]!)
}
