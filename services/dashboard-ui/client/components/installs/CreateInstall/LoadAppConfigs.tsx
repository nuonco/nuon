import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { TApp, TAppConfig } from '@/types'
import { FormSkeleton } from './FormSkeleton'

interface LoadAppConfigsProps {
  app: TApp
  configs: TAppConfig[] | undefined
  isLoading: boolean
  error: any
  onSelectApp: (app: TApp | undefined) => void
  children?: React.ReactNode
}

export const LoadAppConfigs = ({
  app,
  configs,
  isLoading,
  error,
  onSelectApp,
  children,
}: LoadAppConfigsProps) => {
  if (isLoading) {
    return (
      <div>
        <div className="pb-4">
          <Button
            className="cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
            onClick={() => onSelectApp(undefined)}
          >
            <Icon variant="CaretLeftIcon" weight="bold" />
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
            className="cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
            onClick={() => onSelectApp(undefined)}
          >
            <Icon variant="CaretLeftIcon" weight="bold" />
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
            className="cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
            onClick={() => onSelectApp(undefined)}
          >
            <Icon variant="CaretLeftIcon" weight="bold" />
            Back
          </Button>
        </div>
        <Banner theme="error">No configurations found for this app</Banner>
      </div>
    )
  }

  return <>{children}</>
}
