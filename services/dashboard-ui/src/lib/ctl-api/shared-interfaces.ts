// Shared interfaces used across multiple ctl-api modules

export interface IGetApps {
  orgId: string
}

export interface IGetApp {
  appId: string
  orgId: string
}

export interface IGetComponent {
  componentId: string
  orgId: string
}

export interface IGetInstalls {
  orgId: string
}

export interface IGetInstall {
  installId: string
  orgId: string
}

export interface IGetWorkspace {
  orgId: string
  workspaceId: string
}

export interface IGetRunner {
  orgId: string
  runnerId: string
}

export interface IGetOrg {
  orgId: string
}
