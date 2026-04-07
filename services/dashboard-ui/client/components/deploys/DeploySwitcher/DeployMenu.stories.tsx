export default {
  title: 'Deploys/DeployMenu',
}

import { useRef } from 'react'
import { DeployMenu } from './DeployMenu'

const mockDeploys = Array.from({ length: 3 }, (_, i) => ({
  id: `deploy-${i + 1}`,
  created_by: { email: `user${i}@example.com` },
  status_v2: { status: i === 0 ? 'active' : 'success' },
})) as any[]

export const Default = () => {
  const ref = useRef<HTMLDivElement>(null)
  return (
    <DeployMenu
      activeDeployId="deploy-1"
      deploys={mockDeploys}
      isLoading={false}
      hasError={false}
      orgId="org-1"
      installId="install-1"
      componentId="comp-1"
      scrollRef={ref}
      limit={8}
    />
  )
}

export const Empty = () => {
  const ref = useRef<HTMLDivElement>(null)
  return (
    <DeployMenu
      activeDeployId=""
      deploys={[]}
      isLoading={false}
      hasError={false}
      orgId="org-1"
      installId="install-1"
      componentId="comp-1"
      scrollRef={ref}
      limit={8}
    />
  )
}
