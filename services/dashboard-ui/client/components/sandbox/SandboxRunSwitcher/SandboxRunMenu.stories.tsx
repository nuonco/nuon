export default {
  title: 'Sandbox/SandboxRunMenu',
}

import { useRef } from 'react'
import { SandboxRunMenu } from './SandboxRunMenu'

const mockRuns = Array.from({ length: 3 }, (_, i) => ({
  id: `run-${i + 1}`,
  created_by: { email: `user${i}@example.com` },
  status_v2: { status: i === 0 ? 'active' : 'success' },
})) as any[]

export const Default = () => {
  const ref = useRef<HTMLDivElement>(null)
  return (
    <SandboxRunMenu
      activeSandboxRunId="run-1"
      sandboxRuns={mockRuns}
      isLoading={false}
      hasError={false}
      orgId="org-1"
      installId="install-1"
      scrollRef={ref}
      limit={8}
    />
  )
}

export const Empty = () => {
  const ref = useRef<HTMLDivElement>(null)
  return (
    <SandboxRunMenu
      activeSandboxRunId=""
      sandboxRuns={[]}
      isLoading={false}
      hasError={false}
      orgId="org-1"
      installId="install-1"
      scrollRef={ref}
      limit={8}
    />
  )
}
