'use client'

import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import type { TApp, TAppConfig } from '@/types'
import { CreateInstallFromApp } from './CreateInstallFromApp'
import { FormSkeleton } from './FormSkeleton'

interface LoadAppConfigsProps {
  app: TApp
  onSelectApp: (app: TApp | undefined) => void
  onClose: () => void
  formRef?: React.RefObject<HTMLFormElement>
  modalId?: string
  onLoadingChange?: (loading: boolean) => void
}

export const LoadAppConfigs = ({
  app,
  onSelectApp,
  onClose,
  formRef,
  modalId,
  onLoadingChange,
}: LoadAppConfigsProps) => {
  const { org } = useOrg()
  const {
    data: configs,
    isLoading,
    error,
  } = useQuery<TAppConfig[]>({
    path: `/api/orgs/${org?.id}/apps/${app.id}/configs`,
  })

  if (isLoading) {
    return (
      <div>
        <div className="pb-4">
          <Button
            className="!flex items-center gap-1.5 cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
            onClick={() => onSelectApp(undefined)}
          >
            <Icon variant="CaretLeft" weight="bold" />
            Back
          </Button>
        </div>
        <FormSkeleton />
      </div>
    )
  }

  if (error) {
    return (
      <div>
        <div className="pb-4">
          <Button
            className="!flex items-center gap-1.5 cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
            onClick={() => onSelectApp(undefined)}
          >
            <Icon variant="CaretLeft" weight="bold" />
            Back
          </Button>
        </div>
        <Banner theme="error">
          {error.error || 'Unable to load app configs'}
        </Banner>
      </div>
    )
  }

  if (!configs || configs.length === 0) {
    return (
      <div>
        <div className="pb-4">
          <Button
            className="!flex items-center gap-1.5 cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
            onClick={() => onSelectApp(undefined)}
          >
            <Icon variant="CaretLeft" weight="bold" />
            Back
          </Button>
        </div>
        <Banner theme="error">No configurations found for this app</Banner>
      </div>
    )
  }

  return (
    <CreateInstallFromApp
      app={app}
      configId={configs[0].id}
      onSelectApp={onSelectApp}
      onClose={onClose}
      formRef={formRef}
      modalId={modalId}
      onLoadingChange={onLoadingChange}
    />
  )
}