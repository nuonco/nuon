import {
  createContext,
  createElement,
  useContext,
  useMemo,
  type ReactNode,
} from 'react'
import { useMatch, useParams } from 'react-router'
import { useInstall } from '@/hooks/use-install'
import { useNewInstallIA } from '@/hooks/use-new-install-ia'
import { useOrg } from '@/hooks/use-org'
import {
  installBreadcrumbs,
  installHref,
  type TInstallHrefInput,
} from '@/lib/install-path'
import type { TNavLink } from '@/types'

type TInstallHref = (
  input: Omit<TInstallHrefInput, 'nested'>
) => string

type TInstallRouting = {
  nested: boolean
  href: TInstallHref
}

const InstallRoutingContext = createContext<TInstallRouting | null>(null)

const legacyHref: TInstallHref = (input) => installHref(input)

export const InstallRoutingProvider = ({ children }: { children: ReactNode }) => {
  const nested = useNewInstallIA()
  const { org } = useOrg()
  const value = useMemo<TInstallRouting>(
    () => ({
      nested,
      href: (input) =>
        installHref({
          ...input,
          orgId: input.orgId ?? org?.id,
          nested,
        }),
    }),
    [nested, org?.id]
  )

  return createElement(InstallRoutingContext.Provider, { value }, children)
}

export const useInstallHref = () =>
  useContext(InstallRoutingContext)?.href ?? legacyHref

export const useInstallNested = () =>
  useContext(InstallRoutingContext)?.nested ?? false

export const useInstallLink = () => {
  const href = useInstallHref()
  const params = useParams()
  return (input: Omit<TInstallHrefInput, 'nested'>) =>
    href({
      ...input,
      orgId: input.orgId ?? params.orgId,
      appId: input.appId ?? params.appId,
    })
}

export const useInstallRouteMatch = (suffix: string) => {
  const orgMatch = useMatch(`/:orgId/installs/:installId${suffix}`)
  const appMatch = useMatch(
    `/:orgId/apps/:appId/installs/:installId${suffix}`
  )
  const match = orgMatch ?? appMatch
  if (!match) return null
  return {
    ...match,
    params: match.params as Record<string, string | undefined>,
  }
}

export const useInstallPage = () => {
  const { org } = useOrg()
  const installState = useInstall()
  const { install } = installState
  const params = useParams()
  const nested = useInstallNested()
  const appId = install.app_id ?? params.appId
  const href = (suffix = '') =>
    installHref({
      orgId: org?.id,
      appId,
      installId: install.id,
      nested,
      suffix,
    })
  const breadcrumbs = (
    legacy: TNavLink[],
    tail: { suffix: string; text?: string }[] = []
  ) =>
    installBreadcrumbs({
      nested,
      orgId: org?.id,
      orgName: org?.name,
      appId,
      appName: install.app?.name,
      installId: install.id,
      installName: install.name,
      legacy,
      tail,
    })

  return { ...installState, org, appId, nested, href, breadcrumbs }
}
