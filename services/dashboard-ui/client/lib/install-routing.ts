import type { LoaderFunctionArgs, ShouldRevalidateFunction } from 'react-router'
import { redirect } from 'react-router'
import { getInstall } from '@/lib/ctl-api/installs/get-install'
import { getOrg } from '@/lib/ctl-api/orgs/get-org'
import {
  installPathnameSuffix,
  isNewInstallIAEnabled,
} from '@/lib/install-path'
import { queryClient } from '@/lib/query-client'

export const orgHasNewInstallIA = async (orgId: string) => {
  const org = await queryClient.ensureQueryData({
    queryKey: ['org', orgId],
    queryFn: () => getOrg({ orgId }),
  })
  return isNewInstallIAEnabled(org.features)
}

export const resolveInstallPrefix = async (params: {
  orgId?: string
  appId?: string
  installId?: string
}) => {
  const { orgId, appId, installId } = params
  if (!orgId || !installId) return ''
  const nested = await orgHasNewInstallIA(orgId)
  if (!nested) return `/${orgId}/installs/${installId}`
  if (appId) return `/${orgId}/apps/${appId}/installs/${installId}`
  const install = await queryClient.ensureQueryData({
    queryKey: ['install', orgId, installId],
    queryFn: () => getInstall({ orgId, installId }),
  })
  if (!install.app_id) return `/${orgId}/installs/${installId}`
  return `/${orgId}/apps/${install.app_id}/installs/${installId}`
}

const locationTail = (request: Request, installId: string) => {
  const url = new URL(request.url)
  return `${installPathnameSuffix(url.pathname, installId)}${url.search}${url.hash}`
}

export const redirectLegacyInstallRoute = async ({
  params,
  request,
}: LoaderFunctionArgs) => {
  if (!params.orgId || !params.installId) return null
  if (!(await orgHasNewInstallIA(params.orgId))) return null
  const prefix = await resolveInstallPrefix(params)
  if (!prefix.includes('/apps/')) return null
  return redirect(`${prefix}${locationTail(request, params.installId)}`)
}

export const redirectNestedInstallRoute = async ({
  params,
  request,
}: LoaderFunctionArgs) => {
  if (!params.orgId || !params.installId) return null
  const tail = locationTail(request, params.installId)
  if (!(await orgHasNewInstallIA(params.orgId))) {
    return redirect(`/${params.orgId}/installs/${params.installId}${tail}`)
  }
  const install = await queryClient.ensureQueryData({
    queryKey: ['install', params.orgId, params.installId],
    queryFn: () =>
      getInstall({ orgId: params.orgId!, installId: params.installId! }),
  })
  if (!install.app_id || install.app_id === params.appId) return null
  return redirect(
    `/${params.orgId}/apps/${install.app_id}/installs/${params.installId}${tail}`
  )
}

export const redirectWorkflowDetailRoute = async ({
  params,
  request,
}: LoaderFunctionArgs) => {
  if (!params.orgId || !params.installId || !params.workflowId) return null
  const nested = await orgHasNewInstallIA(params.orgId)
  const prefix = await resolveInstallPrefix(params)
  const url = new URL(request.url)
  const tail = `${url.search}${url.hash}`
  const suffix = installPathnameSuffix(url.pathname, params.installId)
  const onDeployments = suffix.startsWith(`/deployments/${params.workflowId}`)
  const onHistory = suffix.startsWith(`/history/${params.workflowId}`)
  if (nested) {
    if (onDeployments) return null
    return redirect(`${prefix}/deployments/${params.workflowId}${tail}`)
  }
  if (onHistory) return null
  return redirect(`${prefix}/history/${params.workflowId}${tail}`)
}

export const installRouteShouldRevalidate: ShouldRevalidateFunction = ({
  currentParams,
  nextParams,
}) =>
  currentParams.orgId !== nextParams.orgId ||
  currentParams.installId !== nextParams.installId ||
  currentParams.appId !== nextParams.appId
