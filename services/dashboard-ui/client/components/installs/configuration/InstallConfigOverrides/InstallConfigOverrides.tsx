import type { ReactNode } from 'react'
import { Skeleton } from '@/components/common/Skeleton'
import { ComponentOverridesList } from '@/components/install-overrides/ComponentOverridesList'
import type { TAppInput, TComponentType } from '@/types'

export interface IInstallConfigOverrides {
  action?: ReactNode
  componentTypes?: Record<string, TComponentType>
  inputs?: TAppInput[]
  isLoading?: boolean
  values?: Record<string, string>
}

export const InstallConfigOverrides = ({
  action,
  componentTypes,
  inputs,
  isLoading,
  values,
}: IInstallConfigOverrides) => {
  if (isLoading) return <Skeleton height="180px" width="100%" />

  return (
    <div className="flex flex-col gap-4">
      {action ? <div className="flex justify-end">{action}</div> : null}
      <ComponentOverridesList
        inputs={inputs}
        values={values}
        codeBlockVariant="viewer"
        componentTypes={componentTypes}
        muteDisabled
        typeVariant="icon"
      />
    </div>
  )
}
