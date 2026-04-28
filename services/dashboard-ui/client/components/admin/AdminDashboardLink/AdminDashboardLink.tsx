import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'

export const AdminDashboardLink = ({
  path,
  label,
}: {
  path: string
  label: string
}) => {
  const { isAdmin } = useAuth()
  const config = useConfig()

  if (!isAdmin || !config.adminDashboardUrl) {
    return null
  }

  const href = `${config.adminDashboardUrl}${path}`

  return (
    <Link className="text-xs" href={href} target="_blank">
      {label} <Icon variant="ArrowSquareOutIcon" />
    </Link>
  )
}
