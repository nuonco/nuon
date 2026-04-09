import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'

interface IAdminDashboardLink {
  path: string
  label: string
  isVisible: boolean
  adminDashboardUrl: string
}

export const AdminDashboardLink = ({
  path,
  label,
  isVisible,
  adminDashboardUrl,
}: IAdminDashboardLink) => {
  if (!isVisible || !adminDashboardUrl) {
    return null
  }

  const href = `${adminDashboardUrl}${path}`

  return (
    <Link className="text-xs" href={href} target="_blank">
      {label} <Icon variant="ArrowSquareOutIcon" />
    </Link>
  )
}
