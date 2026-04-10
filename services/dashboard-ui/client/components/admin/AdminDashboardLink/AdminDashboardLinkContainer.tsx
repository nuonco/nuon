import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { AdminDashboardLink } from './AdminDashboardLink'

export const AdminDashboardLinkContainer = ({
  path,
  label,
}: {
  path: string
  label: string
}) => {
  const { user, isLoading } = useAuth()
  const config = useConfig()
  const isVisible = !isLoading && !!user?.email?.endsWith('@nuon.co')

  return (
    <AdminDashboardLink
      path={path}
      label={label}
      isVisible={isVisible}
      adminDashboardUrl={config.adminDashboardUrl ?? ''}
    />
  )
}
