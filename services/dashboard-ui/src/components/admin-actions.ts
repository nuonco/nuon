'use server'

import type { TRunner } from '@/types'
import { auth0 } from '@/lib/auth'
import { API_URL, ADMIN_API_URL } from '@/configs/api'

async function adminAction(
  domain: string,
  path: string,
  errMessage = 'Admin action failed',
  options: { usePublicAPI?: boolean } = { usePublicAPI: false }
) {
  const { user } = await auth0.getSession()

  try {
    const result = await fetch(
      `${options?.usePublicAPI ? API_URL : ADMIN_API_URL}/v1/${domain}/${path}`,
      {
        method: 'POST',
        body: '{}',
        headers: {
          'Content-Type': 'application/json',
          'X-Nuon-Admin-Email': user?.email,
        },
      }
    ).then((r) => r.json())
    return { status: 201, result }
  } catch (error) {
    throw new Error(errMessage)
  }
}

export async function getToken() {
  const result = await auth0.getAccessToken()
  return { status: 200, result }
}

// org admin actions
// =========================================

async function adminOrgAction(
  orgId: string,
  action: string,
  errMessage = 'Admin org action failed'
) {
  return adminAction('orgs', `${orgId}/${action}`, errMessage)
}

export async function addSupportUsersToOrg(orgId: string) {
  return adminOrgAction(
    orgId,
    'admin-support-users',
    'Failed to add support users to the org'
  )
}

export async function removeSupportUsersFromOrg(orgId: string) {
  return adminOrgAction(
    orgId,
    'admin-remove-support-users',
    'Failed to remove support users from the org'
  )
}

export async function reprovisionOrg(orgId: string) {
  return adminOrgAction(
    orgId,
    'admin-reprovision',
    'Failed to kick off org reprovision'
  )
}

export async function restartOrg(orgId: string) {
  return adminOrgAction(
    orgId,
    'admin-restart',
    'Failed to restart the org event loop'
  )
}

export async function restartOrgRunners(orgId: string) {
  return adminOrgAction(
    orgId,
    'admin-restart-runners',
    'Failed to restart the org children event loops'
  )
}

export async function enableOrgDebugMode(orgId: string) {
  return adminOrgAction(
    orgId,
    'admin-debug-mode',
    'Failed to enable debug mode for this org'
  )
}

export async function updateOrgFeature(
  orgId: string,
  formData: FormData,
  list: Array<string>
) {
  const data = Object.fromEntries(formData)
  const features = data['all']
    ? { all: true }
    : list.reduce((acc, feat) => {
        acc[feat] = data.hasOwnProperty(feat)
        return acc
      }, {})
  const { user } = await auth0.getSession()

  try {
    const result = await fetch(
      `${ADMIN_API_URL}/v1/orgs/${orgId}/admin-features`,
      {
        method: 'PATCH',
        body: JSON.stringify({ features }),
        headers: {
          'Content-Type': 'application/json',
          'X-Nuon-Admin-Email': user?.email,
        },
      }
    ).then((r) => r.json())
    return { status: 201, result }
  } catch (error) {
    throw new Error('Unable to update org features')
  }
}

// app admin actions
// =========================================
async function adminAppAction(
  appId: string,
  action: string,
  errMessage = 'Admin app action failed'
) {
  return adminAction('apps', `${appId}/${action}`, errMessage)
}

export async function restartApp(appId: string) {
  return adminAppAction(
    appId,
    'admin-restart',
    'Failed to restart the app event loop'
  )
}

export async function reprovisionApp(appId: string) {
  return adminAppAction(
    appId,
    'admin-reprovision',
    'Failed to kick off app reprovision'
  )
}

// install admin actions
// =========================================

async function adminInstallAction(
  installId: string,
  action: string,
  errMessage = 'Admin install action failed'
) {
  return adminAction('installs', `${installId}/${action}`, errMessage)
}

export async function reprovisionInstall(installId: string) {
  return adminInstallAction(
    installId,
    'admin-reprovision',
    'Failed to kick off install reprovision'
  )
}

export async function reprovisionInstallRunner(installId: string) {
  return adminInstallAction(
    installId,
    'admin-reprovision-runner',
    'Failed to kick off install runner reprovision'
  )
}

export async function restartInstall(installId: string) {
  return adminInstallAction(
    installId,
    'admin-restart',
    'Failed to restart install'
  )
}

export async function teardownInstallComponents(installId: string) {
  return adminInstallAction(
    installId,
    'admin-teardown-components',
    'Failed to teardown install components'
  )
}

export async function updateInstallSandbox(installId: string) {
  return adminInstallAction(
    installId,
    'admin-update-sandbox',
    'Failed to update install sandbox'
  )
}

export async function restartInstallRunner(installId: string) {
  const runner = await getInstallRunner(installId)
  return adminAction(
    'runners',
    `${runner?.id}/restart`,
    'Failed to restart install runner'
  )
}

export async function shutdownInstallRunnerJob(installId: string) {
  const runner = await getInstallRunner(installId)
  return adminAction(
    'runners',
    `${runner?.id}/shutdown-job`,
    'Failed to kick off install runner shutdown job'
  )
}

export async function getInstallRunner(installId: string): Promise<TRunner> {
  const runner = await fetch(
    `${ADMIN_API_URL}/v1/installs/${installId}/admin-get-runner`
  ).then((r) => r.json())
  return runner
}

export async function getOrgRunner(orgId: string): Promise<TRunner> {
  const runner = await fetch(
    `${ADMIN_API_URL}/v1/orgs/${orgId}/admin-get-runner`
  ).then((r) => r.json())
  return runner
}

export async function restartOrgRunner(orgId: string) {
  const runner = await getOrgRunner(orgId)
  return adminAction(
    'runners',
    `${runner?.id}/restart`,
    'Failed to restart org runner'
  )
}

export async function shutdownOrgRunnerJob(orgId: string) {
  const runner = await getOrgRunner(orgId)
  return adminAction(
    'runners',
    `${runner?.id}/shutdown-job`,
    'Failed to kick off org runner shutdown job'
  )
}

export async function gracefulRunnerShutdown(runnerId: string) {
  return adminAction(
    'runners',
    `${runnerId}/graceful-shutdown`,
    'Failed to kickoff graceful shutdown'
  )
}

export async function forceRunnerShutdown(runnerId: string) {
  return adminAction(
    'runners',
    `${runnerId}/force-shutdown`,
    'Failed to kickoff forced shutdown'
  )
}

export async function invalidateRunnerToken(runnerId: string) {
  return adminAction(
    'runners',
    `${runnerId}/invalidate-service-account-token`,
    'Failed to invalidate service account token'
  )
}

export async function mngShutdownRunner(runnerId: string) {
  return adminAction(
    'runners',
    `${runnerId}/mng/shutdown`,
    'Failed to kickoff mng shutdown',
    { usePublicAPI: true }
  )
}

export async function mngShutdownRunnerVM(runnerId: string) {
  return adminAction(
    'runners',
    `${runnerId}/mng/shutdown-vm`,
    'Failed to kickoff mng shutdown vm',
    { usePublicAPI: true }
  )
}

export async function mngUpdateRunner(runnerId: string) {
  return adminAction(
    'runners',
    `${runnerId}/mng/update`,
    'Failed to kickoff mng update runner',
    { usePublicAPI: true }
  )
}

export async function gracefulInstallRunnerShutdown(installId: string) {
  const runner = await getInstallRunner(installId)
  return gracefulRunnerShutdown(runner.id)
}

export async function forceInstallRunnerShutdown(installId: string) {
  const runner = await getInstallRunner(installId)
  return forceRunnerShutdown(runner.id)
}

export async function invalidateInstallRunnerToken(installId: string) {
  const runner = await getInstallRunner(installId)
  return invalidateRunnerToken(runner.id)
}

export async function gracefulOrgRunnerShutdown(orgId: string) {
  const runner = await getOrgRunner(orgId)
  return gracefulRunnerShutdown(runner.id)
}

export async function forceOrgRunnerShutdown(orgId: string) {
  const runner = await getOrgRunner(orgId)
  return forceRunnerShutdown(runner.id)
}

export async function invalidateOrgRunnerToken(orgId: string) {
  const runner = await getOrgRunner(orgId)
  return invalidateRunnerToken(runner.id)
}
