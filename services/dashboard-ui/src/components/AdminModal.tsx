'use client'

import { useParams } from 'next/navigation'
import React, { type FC, useState } from 'react'
import { useUser } from '@auth0/nextjs-auth0/client'
import { CheckCircle, XCircle, Spinner } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Grid } from '@/components/Grid'
import { Modal } from '@/components/Modal'
import { ToolTip } from '@/components/ToolTip'
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
  restartOrgChildren,
  teardownInstallComponents,
  updateInstallSandbox,
} from '@/components/admin-actions'

type TAdminAction = {
  action: () => Promise<any>
  description: string
  text: string
}

export const AdminModal: FC<{ orgId: string }> = () => {
  const params = useParams()
  const { user } = useUser()
  const [isOpen, setIsOpen] = useState(false)

  const orgActions: Array<TAdminAction> = [
    {
      action: () => addSupportUsersToOrg(params?.['org-id'] as string),
      description: 'Add all nuon support users to current org',
      text: 'Add support users',
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
      action: () => restartOrgChildren(params?.['org-id'] as string),
      description: 'Restart all of current org children event loops',
      text: 'Restart org children',
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
  ]

  return user && /@nuon.co\s*$/.test(user?.email) ? (
    <>
      <Button
        className="text-sm"
        onClick={() => {
          setIsOpen(true)
        }}
        variant="ghost"
      >
        Admin controls
      </Button>
      <Modal
        heading="Admin controls"
        isOpen={isOpen}
        onClose={() => {
          setIsOpen(false)
        }}
      >
        <div className="flex flex-col gap-8 divide-y">
          <div className="flex flex-col gap-4">
            <Text variant="semi-18">Org admin controls</Text>
            <Grid>
              {orgActions.map((action) => (
                <AdminAction key={action.text} {...action} />
              ))}
            </Grid>
          </div>

          {params?.['app-id'] ? (
            <div className="flex flex-col gap-4 pt-4">
              <Text variant="semi-18">App admin controls</Text>
              <Grid>
                {appActions.map((action) => (
                  <AdminAction key={action.text} {...action} />
                ))}
              </Grid>
            </div>
          ) : null}

          {params?.['install-id'] ? (
            <div className="flex flex-col gap-4 pt-4">
              <Text variant="semi-18">Install admin controls</Text>
              <Grid>
                {installActions.map((action) => (
                  <AdminAction key={action.text} {...action} />
                ))}
              </Grid>
            </div>
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

interface IAdminButton {
  children: React.ReactNode
  action: () => Promise<Record<string, any>>
}

const AdminBtn: FC<IAdminButton> = ({ children, action }) => {
  const [actionStatus, setActionStatus] = useState<
    'succeeded' | 'failed' | null
  >(null)
  const [isActing, setIsActing] = useState(false)

  return (
    <Button
      className="flex gap-2 items-center justify-center text-base"
      onClick={() => {
        setIsActing(true)
        action()
          .then((res) => {
            if (res.status === 201) {
              setActionStatus('succeeded')
            }
            setIsActing(false)
          })
          .catch((err) => {
            setActionStatus('failed')
            console.error(err)
            setIsActing(false)
          })
      }}
      disabled={isActing}
    >
      {isActing ? (
        <>
          <Spinner className="animate-spin" /> Executing...
        </>
      ) : (
        <>
          <ActionIcon status={actionStatus} />
          {children}
        </>
      )}
    </Button>
  )
}

const ActionIcon: FC<{ status: 'succeeded' | 'failed' | null }> = ({
  status,
}) => {
  return status ? (
    <ToolTip tipContent={`Admin action ${status}`} isIconHidden>
      {status === 'failed' ? (
        <XCircle className="text-red-500" />
      ) : (
        <CheckCircle className="text-green-500" />
      )}
    </ToolTip>
  ) : null
}
