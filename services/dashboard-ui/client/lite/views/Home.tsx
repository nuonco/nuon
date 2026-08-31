import { useQuery } from '@tanstack/react-query'
import { getMe, getOrgs } from '@/lib'
import { useConfig } from '@/hooks/use-config'
import { ThemeSwitcher } from '../components/ThemeSwitcher'
import { useTheme } from '../hooks/use-theme'

export const Home = () => {
  const config = useConfig()
  const { preference, theme } = useTheme()

  const { data: me, isLoading: isLoadingMe, error: meError } = useQuery({
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

  return (
    <main className="mx-auto flex max-w-2xl flex-col gap-8 px-6 py-16">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-semibold text-primary">Nuon lite</h1>
        <p className="text-sm text-tertiary">
          Lite dashboard shell. Version {config?.version ?? 'dev'}.
        </p>
      </div>

      <div className="flex flex-col gap-3">
        <ThemeSwitcher />
        <p className="text-sm text-tertiary">
          Preference <span className="font-mono text-secondary">{preference}</span>,
          showing <span className="font-mono text-secondary">{theme}</span>.
        </p>
      </div>

      <div className="rounded-xl border border-divider bg-surface-01 p-5">
        {isLoading ? (
          <p className="text-sm text-tertiary">Loading account…</p>
        ) : meError ? (
          <p className="text-sm text-secondary">
            Unable to load the current account.
          </p>
        ) : (
          <dl className="flex flex-col gap-4 text-sm">
            <div className="flex flex-col gap-1">
              <dt className="text-tertiary">Account</dt>
              <dd className="font-mono text-primary">{me?.email}</dd>
            </div>
            <div className="flex flex-col gap-1">
              <dt className="text-tertiary">Orgs</dt>
              <dd className="flex flex-col gap-1 font-mono text-primary">
                {orgs?.length ? (
                  orgs.map((org) => (
                    <span key={org.id}>
                      {org.name} — {org.id}
                    </span>
                  ))
                ) : (
                  <span className="font-sans text-tertiary">No orgs yet</span>
                )}
              </dd>
            </div>
          </dl>
        )}
      </div>
    </main>
  )
}
