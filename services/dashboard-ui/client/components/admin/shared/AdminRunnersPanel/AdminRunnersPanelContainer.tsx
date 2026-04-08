import { useState, useEffect } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { type IPanel } from '@/components/surfaces/Panel'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useAuth } from '@/hooks/use-auth'
import { adminRestartOrgRunners, getInstalls } from '@/lib'
import type { TInstall } from '@/types'
import { AdminRunnersPanel } from './AdminRunnersPanel'

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
  const [installs, setInstalls] = useState<TInstall[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()

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
      const result = await getInstalls({ orgId })
      setInstalls(result.data)
    } catch {
      setError('Unable to load org installs')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchInstalls()
  }, [orgId])

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
      {...props}
    />
  )
}
