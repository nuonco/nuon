import type { TAppInputConfig, TInstall } from '@/types'

export interface ICreateInstallForm {
  appId: string
  platform: 'aws' | 'azure' | 'gcp'
  inputConfig?: TAppInputConfig
  onSubmit?: (formData: FormData) => Promise<any>
  onSuccess?: (result: any) => void
  onCancel: () => void
  isLoading?: boolean
  error?: any
  onRegisterClearDraft?: (clearFn: () => void) => void
}

export interface IUpdateInstallForm {
  install: TInstall
  platform?: 'aws' | 'azure' | 'gcp'
  inputConfig?: TAppInputConfig
  onSubmit?: (formData: FormData) => Promise<any>
  onSuccess?: (result: any) => void
  onCancel: () => void
  isLoading?: boolean
  error?: any
  onFormSubmit?: () => void
  onRegisterClearDraft?: (clearFn: () => void) => void
}

export interface IPlatformFields {
  platform: 'aws' | 'azure' | 'gcp'
  draftValues?: Record<string, string> | null
}

export interface IInputConfigFields {
  inputConfig: TAppInputConfig
  install?: TInstall
  draftValues?: Record<string, string> | null
}

export interface IFieldWrapper {
  children: React.ReactElement
  labelText: string
  helpText?: string
}