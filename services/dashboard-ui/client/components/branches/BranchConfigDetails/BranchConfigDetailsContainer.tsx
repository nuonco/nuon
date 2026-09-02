import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { panelTriggerClass } from '@/components/surfaces/panel-trigger'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getAppConfig } from '@/lib'
import type { TAppConfig } from '@/types'
import {
  BranchConfigDetails,
  type IBranchConfigDetails,
} from './BranchConfigDetails'

interface IBranchConfigDetailsPanel
  extends Omit<IBranchConfigDetails, 'fullConfig' | 'isLoading'> {
  config: TAppConfig
  appId?: string
}

export const BranchConfigDetailsPanel = ({
  config,
  appId,
  ...props
}: IBranchConfigDetailsPanel) => {
  const { org } = useOrg()

  const { data: fullConfig, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-config', org?.id, appId, config?.id, 'recurse'],
    queryFn: () =>
      getAppConfig({
        orgId: org!.id,
        appId: appId!,
        appConfigId: config!.id!,
        recurse: true,
      }),
    enabled: !!org?.id && !!appId && !!config?.id,
  })

  return (
    <BranchConfigDetails
      config={config}
      fullConfig={fullConfig}
      isLoading={isLoading}
      {...props}
    />
  )
}

export const BranchConfigDetailsButton = ({
  config,
  appId,
  children,
  ...props
}: {
  config: TAppConfig
  appId?: string
} & IButtonAsButton) => {
  const { addPanel } = useSurfaces()

  return (
    <Button
      variant="ghost"
      className={panelTriggerClass}
      onClick={() =>
        addPanel(
          <BranchConfigDetailsPanel config={config} appId={appId} />,
          `branch-config-${config?.id}`
        )
      }
      {...props}
    >
      {children}
    </Button>
  )
}
