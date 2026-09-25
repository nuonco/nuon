import type { TInstall } from '@/types'

export const activeAppBranchConnection = (install: TInstall) =>
  install.app_branch_connections?.find((connection) => connection.active)

export const installAppBranchId = (install: TInstall) =>
  activeAppBranchConnection(install)?.app_branch_id ?? install.app_branch_id

export const installAppBranchGroup = (install: TInstall) =>
  activeAppBranchConnection(install)?.app_branch_group ??
  install.app_branch_group
