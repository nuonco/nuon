import { useQuery } from '@tanstack/react-query'
import { getMe, getOrgs } from '@/lib'
import { useConfig } from '@/hooks/use-config'
import { Card } from '../components/atoms/Card'
import { Text } from '../components/atoms/Text'
import { ThemeSwitcher } from '../components/molecules/ThemeSwitcher'
import type { INavItem } from '../components/molecules/NavLink'
import { UserDropdown } from '../components/organisms/UserDropdown'
import { DashboardShell } from '../components/templates/DashboardShell'
import { useTheme } from '../hooks/use-theme'

const PRIMARY_NAV: INavItem[] = [
  {
    href: '/',
    label: 'Dashboard',
    icon: 'HouseIcon',
    shortcut: 'g d',
    end: true,
  },
  {
    href: '/apps',
    label: 'Apps',
    icon: 'AppWindowIcon',
    shortcut: 'g a',
  },
  {
    href: '/installs',
    label: 'Installs',
    icon: 'CubeIcon',
    shortcut: 'g i',
  },
]

const SECONDARY_NAV: INavItem[] = [
  {
    href: '/team',
    label: 'Team',
    icon: 'UsersThreeIcon',
    shortcut: 'g t',
  },
  {
    href: '/settings',
    label: 'Settings',
    icon: 'GearIcon',
    shortcut: 'g s',
  },
  {
    href: 'https://docs.nuon.co',
    label: 'Developer docs',
    icon: 'BookOpenTextIcon',
    external: true,
  },
]

export const Home = () => {
  const config = useConfig()
  const { preference, theme } = useTheme()

  const {
    data: me,
    isLoading: isLoadingMe,
    error: meError,
  } = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: getMe,
    staleTime: Infinity,
    retry: false,
  })

  const { data: orgs, isLoading: isLoadingOrgs } = useQuery({
    queryKey: ['orgs'],
    queryFn: () => getOrgs(),
    retry: false,
  })

  const isLoading = isLoadingMe || isLoadingOrgs
  const organization = orgs?.[0]
  const identity = me?.identities?.[0]
  const user = {
    name: identity?.name,
    email: me?.email,
    picture: identity?.picture,
  }

  return (
    <DashboardShell
      primaryNav={PRIMARY_NAV}
      secondaryNav={SECONDARY_NAV}
      userMenu={
        <UserDropdown
          user={user}
          loading={isLoadingMe}
          signOutHref={`${config.authServiceUrl ?? ''}/logout`}
          stretch
        />
      }
      headerLeading={
        <Text variant="caption" weight="semibold" className="truncate">
          {organization?.name ?? 'Nuon'}
        </Text>
      }
      headerActions={<ThemeSwitcher />}
      statusBar={
        <div className="flex h-8 items-center justify-between gap-4 px-4">
          <Text variant="label" color="secondary">
            {meError ? 'Dashboard connection issue' : 'Dashboard connected'}
          </Text>
          <Text variant="label" color="tertiary">
            Version {config.version ?? 'dev'}
          </Text>
        </div>
      }
    >
      <div className="mx-auto flex w-full max-w-4xl flex-col gap-8 py-12">
        <div className="flex flex-col gap-2">
          <Text as="h1" variant="title" color="primary">
            Nuon lite
          </Text>
          <Text variant="caption" color="tertiary">
            Lite dashboard shell. Version {config?.version ?? 'dev'}.
          </Text>
        </div>

        <div className="flex flex-col gap-3">
          <Text variant="caption" color="tertiary">
            Theme preference{' '}
            <Text family="mono" variant="caption" color="secondary">
              {preference}
            </Text>
            , showing{' '}
            <Text family="mono" variant="caption" color="secondary">
              {theme}
            </Text>
            .
          </Text>
        </div>

        <Card>
          {meError ? (
            <Text variant="body" color="secondary">
              Unable to load the current account.
            </Text>
          ) : (
            <dl className="flex flex-col gap-4">
              <div className="flex flex-col gap-1">
                <Text as="dt" variant="label" color="tertiary">
                  Account
                </Text>
                <Text
                  as="dd"
                  variant="body"
                  family="mono"
                  color="primary"
                  loading={isLoading}
                >
                  {me?.email}
                </Text>
              </div>
              <div className="flex flex-col gap-1">
                <Text as="dt" variant="label" color="tertiary">
                  Orgs
                </Text>
                <dd className="flex flex-col gap-1">
                  {isLoading ? (
                    <Text
                      variant="body"
                      family="mono"
                      loading
                      loadingWidth={24}
                    />
                  ) : orgs?.length ? (
                    orgs.map((org) => (
                      <Text
                        key={org.id}
                        variant="body"
                        family="mono"
                        color="primary"
                      >
                        {org.name} — {org.id}
                      </Text>
                    ))
                  ) : (
                    <Text variant="body" color="tertiary">
                      No orgs yet
                    </Text>
                  )}
                </dd>
              </div>
            </dl>
          )}
        </Card>
      </div>
    </DashboardShell>
  )
}
