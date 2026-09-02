import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getMe, getOrgs } from '@/lib'
import { useConfig } from '@/hooks/use-config'
import { CodeBlock } from '../components/molecules/CodeBlock'
import { Text } from '../components/atoms/Text'
import { ThemeSwitcher } from '../components/molecules/ThemeSwitcher'
import { useTheme } from '../hooks/use-theme'
import { installStateJSON } from '../lib/fixtures/install-state'

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
  const installState = useMemo(() => installStateJSON(), [])

  return (
    <main className="mx-auto flex max-w-4xl flex-col gap-8 px-6 py-16">
      <div className="flex flex-col gap-2">
        <Text as="h1" variant="title" color="primary">
          Nuon lite
        </Text>
        <Text variant="caption" color="tertiary">
          Lite dashboard shell. Version {config?.version ?? 'dev'}.
        </Text>
      </div>

      <div className="flex flex-col gap-3">
        <ThemeSwitcher />
        <Text variant="caption" color="tertiary">
          Preference{' '}
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

      <div className="rounded-xl border border-divider bg-surface-01 p-5">
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
              <Text as="dd" variant="body" family="mono" color="primary" loading={isLoading}>
                {me?.email}
              </Text>
            </div>
            <div className="flex flex-col gap-1">
              <Text as="dt" variant="label" color="tertiary">
                Orgs
              </Text>
              <dd className="flex flex-col gap-1">
                {isLoading ? (
                  <Text variant="body" family="mono" loading loadingWidth={24} />
                ) : orgs?.length ? (
                  orgs.map((org) => (
                    <Text key={org.id} variant="body" family="mono" color="primary">
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
      </div>

      {/* Temporary: a production-sized install state, here to exercise the
          worker pool and virtualization in the real app rather than in Ladle. */}
      <div className="flex flex-col gap-3">
        <Text as="h2" variant="heading" color="primary">
          Install state
        </Text>
        <Text variant="caption" color="tertiary">
          {installState.split('\n').length.toLocaleString()} lines, including one
          line of{' '}
          {Math.max(...installState.split('\n').map((line) => line.length)).toLocaleString()}{' '}
          characters.
        </Text>
        <CodeBlock
          language="json"
          filename="install-state.json"
          value={installState}
          maxHeight={560}
          copy
        />
      </div>
    </main>
  )
}
