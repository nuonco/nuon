import type {
  TAppConfig,
  TBuild,
  TComponentConfig,
  TDeploy,
  TInstallComponent,
} from '@/types'
import type { TComponentOverrideCard } from '@/utils/install-utils'

export type TInstallComponentOutletContext = {
  appConfig?: TAppConfig
  config?: TComponentConfig
  dependentIds: string[]
  installComponent?: TInstallComponent
  installValues?: Record<string, string>
  isDisabled: boolean
  isLoading: boolean
  isLoadingConfig: boolean
  isToggleable: boolean
  latestBuilds?: TBuild[]
  latestDeploy?: TDeploy
  overrideCard?: TComponentOverrideCard<{ name?: string; index?: number }>
  removed: boolean
}
