import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import type { IDropdown } from '@/components/common/Dropdown'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useScrollToBottom } from '@/hooks/use-on-scroll-to-bottom'
import { getComponentDeploys } from '@/lib'
import type { TDeploy } from '@/types'
import { InstallComponentDeploySwitcher } from './InstallComponentDeploySwitcher'

interface IInstallComponentDeploySwitcherContainer
  extends Omit<IDropdown, 'children' | 'id' | 'buttonText'> {
  componentId: string
  deployId: string
}

export const InstallComponentDeploySwitcherContainer = ({
  componentId,
  deployId,
  ...props
}: IInstallComponentDeploySwitcherContainer) => {
  const limit = 8
  const [offset, setOffset] = useState(0)
  const [deploys, setDeploys] = useState<TDeploy[]>([])
  const { install } = useInstall()
  const { org } = useOrg()

  const { data, error, isLoading } = useQuery({
    queryKey: ['component-deploys-switcher', org?.id, install?.id, componentId, offset],
    queryFn: () =>
      getComponentDeploys({
        orgId: org.id,
        installId: install.id,
        componentId,
        limit,
        offset,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  const scrollToBottom = useScrollToBottom({
    onScrollToBottom: () => {
      if (data?.pagination?.hasNext) {
        setOffset((prev) => {
          if (prev === 0) {
            return limit + 1
          } else {
            return prev + limit
          }
        })
      }
    },
  })

  useEffect(() => {
    if (data?.data) {
      setDeploys((prev) => {
        const deployMap = new Map(prev.map((deploy) => [deploy.id, deploy]))
        data.data.forEach((deploy) => deployMap.set(deploy.id, deploy))
        return Array.from(deployMap.values())
      })
      scrollToBottom.reset()
    }
  }, [data])

  return (
    <InstallComponentDeploySwitcher
      componentId={componentId}
      deployId={deployId}
      deploys={deploys}
      isLoading={isLoading}
      hasNext={data?.pagination?.hasNext ?? false}
      orgId={org.id}
      installId={install.id}
      onLoadMore={() => {
        setOffset((prev) => prev === 0 ? limit + 1 : prev + limit)
      }}
      onDeploysLoaded={(newDeploys) => setDeploys(newDeploys)}
      scrollRef={scrollToBottom.elementRef}
      {...props}
    />
  )
}
