import type { TAccount, TInstallActionRun, TWorkflowStep } from '@/types'

export interface IActionRunDetails {
  step?: TWorkflowStep
}

export interface IActionRunHeader {
  actionRun?: TInstallActionRun
  isAdhoc?: boolean
  loading?: boolean
  step?: TWorkflowStep
}

export interface IActionRunMetadata {
  actionRun?: TInstallActionRun
  createdBy?: TAccount
  loading?: boolean
  step?: TWorkflowStep
}

export interface IAdhocActionDetails {
  actionRun: TInstallActionRun
}

export interface IStandardActionSteps {
  actionRun?: TInstallActionRun
  loading?: boolean
}

export interface IActionRunLogs {
  actionRun?: TInstallActionRun
  isAdhoc?: boolean
  loading?: boolean
  step?: TWorkflowStep
}
