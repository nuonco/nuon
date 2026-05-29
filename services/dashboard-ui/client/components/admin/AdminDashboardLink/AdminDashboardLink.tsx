import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { useAuth } from '@/hooks/use-auth'

export const AdminDashboardLink = ({
  path,
}: {
  path: string
  label?: string
}) => {
  const { isAdmin } = useAuth()

  if (!isAdmin) {
    return null
  }

  const href = `/admin/dashboard${path}`

  return (
    <Link
      className="text-xs inline-flex items-center gap-1"
      href={href}
      target="_blank"
    >
      admin <Icon variant="ArrowSquareOutIcon" size="14" />
    </Link>
  )
}
