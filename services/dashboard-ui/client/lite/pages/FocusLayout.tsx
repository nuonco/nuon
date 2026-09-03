import { Outlet } from 'react-router'
import { useConfig } from '@/hooks/use-config'
import { Link } from '../components/atoms/Link'
import { UserDropdown } from '../components/organisms/UserDropdown'
import { FocusShell } from '../components/templates/FocusShell'
import { useCurrentUser } from '../hooks/use-current-user'

export const FocusLayout = () => {
  const config = useConfig()
  const { user, isLoading } = useCurrentUser()

  return (
    <FocusShell
      actions={
        <>
          <Link href="https://docs.nuon.co" external variant="caption">
            Developer docs
          </Link>
          <UserDropdown
            user={user}
            loading={isLoading}
            signOutHref={`${config.authServiceUrl ?? ''}/logout`}
          />
        </>
      }
    >
      <Outlet />
    </FocusShell>
  )
}
