'use client'

import { usePathname, useRouter } from 'next/navigation'
import { useState } from 'react'
import { createPortal } from 'react-dom'
import { PencilSimpleLineIcon } from '@phosphor-icons/react'
import { updateInstall } from '@/actions/installs/update-install'
import { updateInstallInputs } from '@/actions/installs/update-install-inputs'
import { Button } from '@/components/Button'
import { InstallForm } from '@/components/InstallForm'
import { Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import type { TAppConfig } from '@/types'
import { ConfirmUpdateModal } from './ConfirmUpdateModal'

export const EditModal = () => {
  const { install } = useInstall()
  const [isOpen, setIsOpen] = useState(false)
  const [isConfirmOpen, setIsConfirmOpen] = useState(false)

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-5xl"
              isOpen={isOpen}
              heading={`Edit install settings?`}
              onClose={() => {
                setIsOpen(false)
              }}
              contentClassName="px-0 py-0"
            >
              <EditForm
                onClose={() => {
                  setIsOpen(false)
                }}
              />
            </Modal>,
            document.body
          )
        : null}
      <ConfirmUpdateModal
        install={install}
        isOpen={isConfirmOpen}
        onClose={(isConfirmed) => {
          setIsOpen(isConfirmed)
          setIsConfirmOpen(false)
        }}
      />
      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full"
        variant="ghost"
        onClick={() => {
          setIsConfirmOpen(true)
        }}
      >
        <PencilSimpleLineIcon size="16" />
        Edit inputs
      </Button>
    </>
  )
}

const EditForm = ({ onClose }: { onClose: () => void }) => {
  const path = usePathname()
  const router = useRouter()
  const { org } = useOrg()
  const { install } = useInstall()

  const {
    data: config,
    isLoading,
    error,
  } = useQuery({
    path: `/api/orgs/${org.id}/apps/${install?.app_id}/configs/${install?.app_config_id}?recurse=true`,
  })

  return (
    <>
      {isLoading ? (
        <div className="p-6">
          <Loading loadingText="Loading configs..." variant="stack" />
        </div>
      ) : error?.error ? (
        <div className="p-6">
          <Notice>{error?.error || 'Unable to load app config'}</Notice>
        </div>
      ) : (
        <InstallForm
          onSubmit={(formData) => {
            const res = updateInstallInputs({
              installId: install.id,
              orgId: org.id,
              formData,
              path,
            })

            if (install?.metadata?.managed_by === 'nuon/cli/install-config') {
              updateInstall({
                installId: install.id,
                managedBy: 'nuon/dashboard',
                orgId: org.id,
              })
            }

            return res
          }}
          onSuccess={({ error, headers, status }) => {
            if (!error && status === 200) {
              router.push(
                `/${org.id}/installs/${install?.id}/workflows/${headers?.['x-nuon-install-workflow-id']}`
              )
            }
          }}
          onCancel={onClose}
          inputConfig={{
            ...config.input,
            input_groups: nestInputsUnderGroups(
              config.input?.input_groups,
              config.input?.inputs
            ),
          }}
          install={install}
        />
      )}
    </>
  )
}

function nestInputsUnderGroups(
  groups: TAppConfig['input']['input_groups'],
  inputs: TAppConfig['input']['inputs']
) {
  return groups
    ? groups.map((group) => ({
        ...group,
        app_inputs: inputs.filter((input) => input.group_id === group.id),
      }))
    : []
}
