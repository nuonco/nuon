import type { TBadgeTheme } from '@/components/common/Badge'
import type { THelmK8sChangeAction, TTerraformChangeAction } from '@/types'

// terraform style utils
export const TERRAFORM_ACTION_BADGE_THEME: Record<
  TTerraformChangeAction,
  TBadgeTheme
> = {
  read: 'info',
  replace: 'brand',
  create: 'success',
  delete: 'error',
  update: 'warn',
  noop: 'neutral',
} as const

export function getTerraformActionBgColor(
  action: TTerraformChangeAction
): string {
  switch (action) {
    case 'read':
      return [
        'bg-blue-100 dark:bg-blue-500/10',
        'hover:!bg-blue-200 dark:hover:!bg-blue-500/20',
        'focus:!bg-blue-200 dark:focus:!bg-blue-500/20',
        'active:!bg-blue-300 dark:active:!bg-blue-500/30',
      ].join(' ')
    case 'replace':
      return [
        'bg-primary-100 dark:bg-primary-500/10',
        'hover:!bg-primary-200 dark:hover:!bg-primary-500/20',
        'focus:!bg-primary-200 dark:focus:!bg-primary-500/20',
        'active:!bg-primary-300 dark:active:!bg-primary-500/30',
      ].join(' ')
    case 'create':
      return [
        'bg-green-100 dark:bg-green-500/10',
        'hover:!bg-green-200 dark:hover:!bg-green-500/20',
        'focus:!bg-green-200 dark:focus:!bg-green-500/20',
        'active:!bg-green-300 dark:active:!bg-green-500/30',
      ].join(' ')
    case 'update':
      return [
        'bg-orange-100 dark:bg-orange-500/10',
        'hover:!bg-orange-200 dark:hover:!bg-orange-500/20',
        'focus:!bg-orange-200 dark:focus:!bg-orange-500/20',
        'active:!bg-orange-300 dark:active:!bg-orange-500/30',
      ].join(' ')
    case 'delete':
      return [
        'bg-red-100 dark:bg-red-500/10',
        'hover:!bg-red-200 dark:hover:!bg-red-500/20',
        'focus:!bg-red-200 dark:focus:!bg-red-500/20',
        'active:!bg-red-300 dark:active:!bg-red-500/30',
      ].join(' ')
    default:
      return [
        'bg-cool-grey-100 dark:bg-dark-grey-500/10',
        'hover:!bg-cool-grey-200 dark:hover:!bg-dark-grey-500/20',
        'focus:!bg-cool-grey-200 dark:focus:!bg-dark-grey-500/20',
        'active:!bg-cool-grey-300 dark:active:!bg-dark-grey-500/30',
      ].join(' ')
  }
}

export function getTerraformActionBorderColor(
  action: TTerraformChangeAction
): string {
  switch (action) {
    case 'read':
      return '!border-l-blue-400 dark:!border-l-blue-600'
    case 'replace':
      return '!border-l-primary-400 dark:!border-l-primary-600'
    case 'create':
      return '!border-l-green-400 dark:!border-l-green-600'
    case 'update':
      return '!border-l-orange-400 dark:!border-l-orange-600'
    case 'delete':
      return '!border-l-red-400 dark:!border-l-red-600'
    default:
      return '!border-l-cool-grey-400 dark:!border-l-cool-grey-500'
  }
}

// helm / k8s style utils
export const HELM_ACTION_BADGE_THEME: Record<
  THelmK8sChangeAction,
  TBadgeTheme
> = {
  add: 'success',
  added: 'success',
  change: 'warn',
  changed: 'warn',
  destroy: 'error',
  destroyed: 'error',
} as const

export function getHelmActionBgColor(action: THelmK8sChangeAction): string {
  switch (action) {
    case 'add':
      return [
        'bg-green-100 dark:bg-green-500/10',
        'hover:!bg-green-200 dark:hover:!bg-green-500/20',
        'focus:!bg-green-200 dark:focus:!bg-green-500/20',
        'active:!bg-green-300 dark:active:!bg-green-500/30',
      ].join(' ')
    case 'added':
      return [
        'bg-green-100 dark:bg-green-500/10',
        'hover:!bg-green-200 dark:hover:!bg-green-500/20',
        'focus:!bg-green-200 dark:focus:!bg-green-500/20',
        'active:!bg-green-300 dark:active:!bg-green-500/30',
      ].join(' ')
    case 'changed':
      return [
        'bg-orange-100 dark:bg-orange-500/10',
        'hover:!bg-orange-200 dark:hover:!bg-orange-500/20',
        'focus:!bg-orange-200 dark:focus:!bg-orange-500/20',
        'active:!bg-orange-300 dark:active:!bg-orange-500/30',
      ].join(' ')
    case 'change':
      return [
        'bg-orange-100 dark:bg-orange-500/10',
        'hover:!bg-orange-200 dark:hover:!bg-orange-500/20',
        'focus:!bg-orange-200 dark:focus:!bg-orange-500/20',
        'active:!bg-orange-300 dark:active:!bg-orange-500/30',
      ].join(' ')
    case 'destroy':
      return [
        'bg-red-100 dark:bg-red-500/10',
        'hover:!bg-red-200 dark:hover:!bg-red-500/20',
        'focus:!bg-red-200 dark:focus:!bg-red-500/20',
        'active:!bg-red-300 dark:active:!bg-red-500/30',
      ].join(' ')
    case 'destroyed':
      return [
        'bg-red-100 dark:bg-red-500/10',
        'hover:!bg-red-200 dark:hover:!bg-red-500/20',
        'focus:!bg-red-200 dark:focus:!bg-red-500/20',
        'active:!bg-red-300 dark:active:!bg-red-500/30',
      ].join(' ')
    default:
      return [
        'bg-cool-grey-100 dark:bg-dark-grey-500/10',
        'hover:!bg-cool-grey-200 dark:hover:!bg-dark-grey-500/20',
        'focus:!bg-cool-grey-200 dark:focus:!bg-dark-grey-500/20',
        'active:!bg-cool-grey-300 dark:active:!bg-dark-grey-500/30',
      ].join(' ')
  }
}

export function getHelmActionBorderColor(action: THelmK8sChangeAction): string {
  switch (action) {
    case 'add':
      return '!border-l-green-400 dark:!border-l-green-600'
    case 'added':
      return '!border-l-green-400 dark:!border-l-green-600'
    case 'changed':
      return '!border-l-orange-400 dark:!border-l-orange-600'
    case 'change':
      return '!border-l-orange-400 dark:!border-l-orange-600'
    case 'destroy':
      return '!border-l-red-400 dark:!border-l-red-600'
    case 'destroyed':
      return '!border-l-red-400 dark:!border-l-red-600'
    default:
      return '!border-l-cool-grey-400 dark:!border-l-cool-grey-500'
  }
}
