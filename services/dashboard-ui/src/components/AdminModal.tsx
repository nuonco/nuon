'use client'

import { useParams } from 'next/navigation'
import React, { type FC, useState } from 'react'
import { useUser } from '@auth0/nextjs-auth0'
import { Gear } from '@phosphor-icons/react'
import { AdminOrgActions } from '@/components/AdminOrgActions'
import { AdminInstallActions } from '@/components/AdminInstallActions'
import { AdminOrgFeatures } from '@/components/AdminOrgFeatures'
import { AdminRunnerModal } from '@/components/AdminRunnerModal'
import { AdminTemporalLink } from '@/components/AdminTemporalLink'
import { AdminBtn } from '@/components/AdminActionButton'
import { Button } from '@/components/Button'
import { Grid } from '@/components/Grid'
import { Modal } from '@/components/Modal'
import { useOrg } from '@/components/Orgs'
import { Text } from '@/components/Typography'
import {
  addSupportUsersToOrg,
  reprovisionApp,
  reprovisionInstall,
  reprovisionInstallRunner,
  reprovisionOrg,
  restartApp,
  restartInstall,
  restartOrg,
  restartOrgRunners,
  restartOrgRunner,
  removeSupportUsersFromOrg,
  shutdownInstallRunnerJob,
  teardownInstallComponents,
  updateInstallSandbox,
  gracefulInstallRunnerShutdown,
  forceInstallRunnerShutdown,
  gracefulOrgRunnerShutdown,
  forceOrgRunnerShutdown,
  enableOrgDebugMode,
  invalidateInstallRunnerToken,
  invalidateOrgRunnerToken,
} from '@/components/admin-actions'

type TAdminAction = {
  action: () => Promise<any>
  description: string
  text: string
}

export const AdminModal: FC<{
  isSidebarOpen: boolean
  isModalOpen?: string
}> = ({ isSidebarOpen, isModalOpen }) => {
  const params = useParams()
  const { user } = useUser()
  const { org } = useOrg()
  const [isOpen, setIsOpen] = useState(Boolean(isModalOpen))

  const orgActions: Array<TAdminAction> = [
    {
      action: () => addSupportUsersToOrg(params?.['org-id'] as string),
      description: 'Add all nuon support users to current org',
      text: 'Add support users',
    },
    {
      action: () => removeSupportUsersFromOrg(params?.['org-id'] as string),
      description: 'Remove all nuon support users from current org',
      text: 'Remove support users',
    },
    {
      action: () => reprovisionOrg(params?.['org-id'] as string),
      description: 'Reprovision current org',
      text: 'Reprovision org',
    },
    {
      action: () => restartOrg(params?.['org-id'] as string),
      description: 'Restart current org event loop',
      text: 'Restart org',
    },
    {
      action: () => restartOrgRunners(params?.['org-id'] as string),
      description: 'Restart all of current org runners',
      text: 'Restart all runners',
    },
    {
      action: () => restartOrgRunner(params?.['org-id'] as string),
      description: 'Restart the current org runner',
      text: 'Restart runner',
    },
    {
      action: () => gracefulOrgRunnerShutdown(params?.['org-id'] as string),
      description: 'Graceful shutdown of current org runner',
      text: 'Graceful org shutdown runner',
    },
    {
      action: () => forceOrgRunnerShutdown(params?.['org-id'] as string),
      description: 'Forceful shutdown of current org runner',
      text: 'Force org shutdown runner',
    },
    {
      action: () => invalidateOrgRunnerToken(params?.['org-id'] as string),
      description:
        'Invalidate a runner service account token, meaning that any live runners will no longer be able to connect to the API.',
      text: 'Invalidate org runner token',
    },
    {
      action: () => enableOrgDebugMode(params?.['org-id'] as string),
      description: 'Debug mode logs all requests for an org.',
      text: 'Enable debug mode',
    },
  ]

  const appActions: Array<TAdminAction> = [
    {
      action: () => reprovisionApp(params?.['app-id'] as string),
      description: 'Reprovision current app',
      text: 'Reprovision app',
    },
    {
      action: () => restartApp(params?.['app-id'] as string),
      description: 'Restart current app event loop',
      text: 'Restart app',
    },
  ]

  const installActions: Array<TAdminAction> = [
    {
      action: () => reprovisionInstall(params?.['install-id'] as string),
      description: 'Reprovision current install sandbox and runner',
      text: 'Reprovision install',
    },
    {
      action: () => reprovisionInstallRunner(params?.['install-id'] as string),
      description: 'Reprovision current install runner',
      text: 'Reprovision runner',
    },
    {
      action: () => restartInstall(params?.['install-id'] as string),
      description: 'Restart current install event loop',
      text: 'Restart install',
    },
    {
      action: () => teardownInstallComponents(params?.['install-id'] as string),
      description: 'Teardown all components on install',
      text: 'Teardown components',
    },
    {
      action: () => updateInstallSandbox(params?.['install-id'] as string),
      description: 'Update install sandbox to the current app sandbox version',
      text: 'Update sandbox',
    },
    /* {
     *   action: () => restartInstallRunner(params?.['install-id'] as string),
     *   description: 'Restart the current install runner',
     *   text: 'Restart runner',
     * }, */
    {
      action: () => shutdownInstallRunnerJob(params?.['install-id'] as string),
      description: 'Shutdown the current install runner job',
      text: 'Shutdown runner job',
    },
    {
      action: () =>
        gracefulInstallRunnerShutdown(params?.['install-id'] as string),
      description: 'Graceful shutdown of current install runner',
      text: 'Graceful install runner shutdown',
    },
    {
      action: () =>
        forceInstallRunnerShutdown(params?.['install-id'] as string),
      description: 'Forceful shutdown of current install runner',
      text: 'Force install runner shutdown',
    },
    {
      action: () =>
        invalidateInstallRunnerToken(params?.['install-id'] as string),
      description:
        'Invalidate a runner service account token, meaning that any live runners will no longer be able to connect to the API.',
      text: 'Invalidate install runner token',
    },
  ]

  return user && /@nuon.co\s*$/.test(user?.email) ? (
    <>
      <Button
        className="text-sm !font-medium flex items-center justify-center gap-2 w-full"
        onClick={() => {
          setIsOpen(true)
        }}
        variant="ghost"
      >
        <span className="inline-block w-[18px] h-[18px]">
          <Gear size={18} />
        </span>{' '}
        {isSidebarOpen ? (
          <span className="text-nowrap truncate">Admin controls</span>
        ) : null}
      </Button>
      <Modal
        heading="Admin controls"
        actions={<AdminRunnerModal />}
        isOpen={isOpen}
        onClose={() => {
          setIsOpen(false)
        }}
      >
        <div className="flex flex-col gap-8 divide-y">
          <AdminOrgActions orgId={params?.['org-id'] as string}>
            <Grid>
              {orgActions.map((action) => (
                <AdminAction key={action.text} {...action} />
              ))}
              <AdminOrgFeatures org={org} />
              <AdminRunnerModal showText />
            </Grid>
          </AdminOrgActions>

          {params?.['app-id'] ? (
            <div className="flex flex-col gap-4 pt-4">
              <Text variant="semi-18">App admin controls</Text>
              <AdminTemporalLink
                namespace="apps"
                id={params?.['app-id'] as string}
              />
              <Grid>
                {appActions.map((action) => (
                  <AdminAction key={action.text} {...action} />
                ))}
              </Grid>
            </div>
          ) : null}

          {params?.['install-id'] ? (
            <AdminInstallActions installId={params?.['install-id'] as string}>
              <Grid>
                {installActions.map((action) => (
                  <AdminAction key={action.text} {...action} />
                ))}
              </Grid>
            </AdminInstallActions>
          ) : null}
        </div>
      </Modal>
    </>
  ) : null
}

const AdminAction: FC<{ action: any; description: string; text: string }> = ({
  action,
  description,
  text,
}) => {
  return (
    <div className="flex flex-col gap-2">
      <Text variant="reg-14">{description}</Text>
      <AdminBtn action={action}>{text}</AdminBtn>
    </div>
  )
}
