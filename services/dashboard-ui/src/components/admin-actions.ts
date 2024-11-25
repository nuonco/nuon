'use server'

import { ADMIN_API_URL } from '@/utils'

async function adminAction(
  domain: string,
  path: string,
  errMessage = 'Admin action failed'
) {
  try {
    const result = await fetch(`${ADMIN_API_URL}/v1/${domain}/${path}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    }).then((r) => r.json())
    return { status: 201, result }
  } catch (error) {
    throw new Error(errMessage)
  }
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

export async function restartOrgChildren(orgId: string) {
  return adminOrgAction(
    orgId,
    'admin-restart-children',
    'Failed to restart the org children event loops'
  )
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
