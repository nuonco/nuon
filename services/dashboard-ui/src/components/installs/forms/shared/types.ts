import type { TAppInputConfig, TInstall } from '@/types'

export interface ICreateInstallForm {
  appId: string
  platform: 'aws' | 'azure'
  inputConfig?: TAppInputConfig
  onSubmit?: (formData: FormData) => Promise<any>
  onSuccess: (result: any) => void
  onCancel: () => void
  isLoading?: boolean
  error?: any
}

export interface IUpdateInstallForm {
  install: TInstall
  platform?: 'aws' | 'azure'
  inputConfig?: TAppInputConfig
  onSubmit?: (formData: FormData) => Promise<any>
  onSuccess: (result: any) => void
  onCancel: () => void
  isLoading?: boolean
  error?: any
  onFormSubmit?: () => void
}

export interface IPlatformFields {
  platform: 'aws' | 'azure'
}

export interface IInputConfigFields {
  inputConfig: TAppInputConfig
  install?: TInstall
}

export interface IFieldWrapper {
  children: React.ReactElement
  labelText: string
  helpText?: string
}