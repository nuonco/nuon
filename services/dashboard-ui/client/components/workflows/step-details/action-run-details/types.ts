import type {
  TAccount,
  TInstallActionRun,
  TWorkflowStep,
} from '@/types'

export interface IActionRunDetails {
  step?: TWorkflowStep
}

export interface IActionRunHeader {
  actionRun: TInstallActionRun
  isAdhoc: boolean
  step?: TWorkflowStep
}

export interface IActionRunMetadata {
  actionRun: TInstallActionRun
  createdBy?: TAccount
  step?: TWorkflowStep
}

export interface IAdhocActionDetails {
  actionRun: TInstallActionRun
}

export interface IStandardActionSteps {
  actionRun: TInstallActionRun
}

export interface IActionRunLogs {
  actionRun: TInstallActionRun
  isAdhoc: boolean
  step?: TWorkflowStep
}
