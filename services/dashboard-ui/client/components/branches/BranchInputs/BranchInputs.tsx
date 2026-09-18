import { AppInputs } from '@/components/apps/config/AppInputs'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Loading } from '@/components/common/Loading'
import type { TAppConfig } from '@/types'

export interface IBranchInputs {
  appConfig?: TAppConfig
  isLoading?: boolean
  isError?: boolean
  isMissingConfig?: boolean
}

export const BranchInputs = ({
  appConfig,
  isLoading = false,
  isError = false,
  isMissingConfig = false,
}: IBranchInputs) => {
  if (isLoading) {
    return (
      <div className="flex min-h-48 items-center justify-center">
        <Loading variant="large" />
      </div>
    )
  }

  if (isError) {
    return (
      <EmptyState
        variant="diagram"
        emptyTitle="Inputs failed"
        emptyMessage="Unable to load the inputs for this branch."
      />
    )
  }

  if (isMissingConfig) {
    return (
      <EmptyState
        variant="diagram"
        emptyTitle="No app config yet"
        emptyMessage="The latest branch update did not produce an app config. Trigger a new run from Updates to try again."
      />
    )
  }

  if (!appConfig?.input?.input_groups?.length) {
    return (
      <EmptyState
        variant="diagram"
        emptyTitle="No app inputs configured"
        emptyMessage="Configure app inputs in this branch's application configuration to see them here."
      />
    )
  }

  return <AppInputs appConfig={appConfig} />
}
