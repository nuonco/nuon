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
    <Link href={href} isExternal>
      admin
    </Link>
  )
}
