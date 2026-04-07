import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { type IMenu } from '@/components/common/Menu'
import { useInstall } from '@/hooks/use-install'
import { useScrollToBottom } from '@/hooks/use-on-scroll-to-bottom'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { getComponentDeploys } from '@/lib'
import type { TDeploy } from '@/types'
import { DeployMenu } from './DeployMenu'

interface IDeployMenuContainer extends Omit<IMenu, 'children'> {
  activeDeployId: string
  componentId: string
}

export const DeployMenuContainer = ({ activeDeployId, componentId }: IDeployMenuContainer) => {
  const limit = 8
  const [offset, setOffset] = useState(0)
  const [deploys, setDeploys] = useState<TDeploy[]>([])
  const { install } = useInstall()
  const { org } = useOrg()
  const queryParams = useQueryParams({
    limit,
    offset,
  })

  const { data, error, isLoading } = useQuery({
    queryKey: ['component-deploys-menu', org?.id, install?.id, componentId, queryParams],
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
    const newDeploys = data?.data ?? []
    setDeploys((prev) => {
      const deployMap = new Map(prev.map((deploy) => [deploy.id, deploy]))
      newDeploys.forEach((deploy) => deployMap.set(deploy.id, deploy))
      return Array.from(deployMap.values())
    })
    scrollToBottom.reset()
  }, [data])

  return (
    <DeployMenu
      activeDeployId={activeDeployId}
      deploys={deploys}
      isLoading={isLoading}
      hasError={!!error}
      orgId={org.id}
      installId={install.id}
      componentId={componentId}
      scrollRef={scrollToBottom.elementRef}
      limit={limit}
    />
  )
}
