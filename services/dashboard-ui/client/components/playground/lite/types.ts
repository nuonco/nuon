import type { TIconVariant } from './Block'

export interface INavItem {
  label: string
  path: string
  icon?: TIconVariant
}

export interface ICrumb {
  label: string
  path?: string
}

export interface IDetailCrumb {
  label: string
  slug?: string
}
