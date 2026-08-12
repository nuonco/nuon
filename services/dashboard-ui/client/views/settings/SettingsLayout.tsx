import { Outlet } from 'react-router'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { SubNav } from '@/components/navigation/SubNav'
import { useOrg } from '@/hooks/use-org'
import { useCLIConfig } from '@/hooks/use-cli-config'
import { PageSidebarProvider } from '@/providers/page-sidebar-provider'
import type { TNavItem } from '@/types/dashboard.types'

export const SettingsLayout = () => {
  return (
    <PageSidebarProvider>
      <SettingsTemplate />
    </PageSidebarProvider>
  )
}

const SettingsTemplate = () => {
  const { org } = useOrg()
  const { data: cliConfig } = useCLIConfig()

  if (!org) return null

  const hasServiceAccountsAndTokens =
    !!org?.features?.['service-accounts-and-tokens']
  const hasSlack = !!org?.features?.['slack']
  const hasTriggers = !!org?.features?.['triggers']
  const hasOIDCFederation = !!cliConfig?.oidc_federation_enabled

  const navLinks = [
    {
      path: `/vcs`,
      iconVariant: 'GitHub' as const,
      text: 'VCS connections',
    },
    {
      path: `/webhooks`,
      iconVariant: 'WebhooksLogoIcon' as const,
      text: 'Webhooks',
    },
    hasSlack && {
      path: `/slack`,
      iconVariant: 'SlackLogoIcon' as const,
      text: 'Slack',
    },
    hasTriggers && {
      path: `/triggers`,
      iconVariant: 'LightningIcon' as const,
      text: 'Triggers',
    },
    hasServiceAccountsAndTokens && {
      path: `/roles`,
      iconVariant: 'ShieldCheckIcon' as const,
      text: 'Roles',
    },
    hasServiceAccountsAndTokens && {
      path: `/api-tokens`,
      iconVariant: 'KeyIcon' as const,
      text: 'API tokens',
    },
    hasServiceAccountsAndTokens && {
      path: `/service-accounts`,
      iconVariant: 'RobotIcon' as const,
      text: 'Service accounts',
    },
    hasOIDCFederation && {
      path: `/oidc`,
      iconVariant: 'ShieldCheckIcon' as const,
      text: 'OIDC federation',
    },
  ].filter(Boolean) as TNavItem[]

  return (
    <PageLayout>
      <PageContent className="border-t" variant="row">
        <SubNav basePath={`/${org?.id}/settings`} links={navLinks} />
        <div className="flex flex-col flex-1 min-w-0">
          <Outlet />
        </div>
      </PageContent>
    </PageLayout>
  )
}
