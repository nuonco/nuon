import { useState, useEffect } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { type IPanel } from '@/components/surfaces/Panel'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useAuth } from '@/hooks/use-auth'
import { adminRestartOrgRunners, getInstalls } from '@/lib'
import type { TInstall } from '@/types'
import type { TPaginationMeta } from '@/lib/api'
import { AdminRunnersPanel } from './AdminRunnersPanel'

const PAGE_SIZE = 10
const PARAM = 'runners_offset'

interface AdminRunnersPanelContainerProps extends IPanel {
  orgId: string
}

export const AdminRunnersPanelContainer = ({
  orgId,
  ...props
}: AdminRunnersPanelContainerProps) => {
  const { org } = useOrg()
  const { addToast } = useToast()
  const { user } = useAuth()
  const adminEmail = user?.email ?? ''
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get(PARAM) || 0)
  const [installs, setInstalls] = useState<TInstall[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()
  const [pagination, setPagination] = useState<TPaginationMeta>({
    hasNext: false,
    offset: 0,
    limit: PAGE_SIZE,
  })

  const { mutate: restartAll, isPending: isRestarting } = useMutation({
    mutationFn: () => adminRestartOrgRunners({ orgId, adminEmail }),
    onSuccess: async () => {
      addToast(
        <Toast heading="Runners Restarted" theme="success">
          <Text>All runners restarted successfully</Text>
        </Toast>
      )
      await fetchInstalls()
    },
    onError: () => {
      addToast(
        <Toast heading="Restart Failed" theme="error">
          <Text>Failed to restart runners. Please try again.</Text>
        </Toast>
      )
    },
  })

  const fetchInstalls = async () => {
    setIsLoading(true)
    setError(undefined)
    try {
      const result = await getInstalls({ orgId, limit: PAGE_SIZE, offset })
      setInstalls(result.data)
      setPagination(result.pagination)
    } catch {
      setError('Unable to load org installs')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchInstalls()
  }, [orgId, offset])

  return (
    <AdminRunnersPanel
      orgId={orgId}
      orgName={org?.name}
      orgRunners={org?.runner_group?.runners ?? []}
      installs={installs}
      isLoading={isLoading}
      error={error}
      isRestarting={isRestarting}
      onRestartAll={() => restartAll()}
      onRefreshInstalls={fetchInstalls}
      pagination={pagination}
      {...props}
    />
  )
}
