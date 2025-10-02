'use client'

import { usePathname } from 'next/navigation'
import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { FaGithub } from 'react-icons/fa'
import { XCircleIcon } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Input } from '@/components/Input'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { removeVCSConnection } from '@/actions/vcs-connection/remove-vcs-connection'
import { useServerAction } from '@/hooks/use-server-action'
import type { TVCSConnection } from '@/types'
import { useOrg } from '@/hooks/use-org'

export const RemoveVCSConnection = ({
  connection,
}: {
  connection: TVCSConnection
}) => {
  const path = usePathname()
  const { org } = useOrg()
  const [confirm, setConfirm] = useState<string>('')
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [isOpen, setIsOpen] = useState(false)

  const connectionName =
    connection?.github_account_name || connection?.github_install_id
  const { data, error, isLoading, execute, status } = useServerAction({
    action: removeVCSConnection,
  })

  const handleClose = () => {
    setIsKickedOff(false)
    setIsOpen(false)
  }

  useEffect(() => {
    if (status === 204) {
      handleClose()
    }
  }, [data, error, status])

  return (
    <>
      <Button
        aria-label="Remove GitHub connection"
        className="ml-auto !p-1 text-red-600 dark:text-red-400"
        variant="ghost"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <XCircleIcon size="16" />
      </Button>

      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-xl"
              onClose={handleClose}
              isOpen={isOpen}
              heading={
                <Text
                  variant="med-14"
                  className="text-red-600 dark:text-red-400"
                >
                  Are you sure you want to disconnect this GitHub account?
                </Text>
              }
            >
              <div className="flex flex-col gap-4 mb-6">
                {error ? <Notice>{error?.error}</Notice> : null}
                <Text variant="med-18">
                  GitHub connection:
                  <span className="ml-2 flex items-center gap-1 text-red-600 dark:text-red-400">
                    <FaGithub className="text-lg" />
                    {connectionName}
                  </span>
                </Text>
                <div className="flex flex-col gap-2">
                  <Text variant="med-12">
                    What happens when you disconnect:
                  </Text>
                  <ul className="list-disc pl-4 flex flex-col gap-2">
                    <li>
                      <Text>
                        Your Nuon organization will lose access to private
                        repositories
                      </Text>
                    </li>
                    <li>
                      <Text>
                        Any workflows using private repos may be affected
                      </Text>
                    </li>
                    <li>
                      <Text>You can reconnect this account at any time</Text>
                    </li>
                  </ul>
                </div>
                <Text>
                  This action cannot be undone, but you can reconnect later.
                </Text>
              </div>
              <div className="w-full">
                <label className="flex flex-col gap-1.5 w-full">
                  <Text variant="med-14">
                    To verify, type{' '}
                    <span className="text-red-800 dark:text-red-500 mx-1">
                      {connectionName}
                    </span>{' '}
                    below
                  </Text>
                  <Input
                    placeholder="GitHub connection name"
                    className="w-full !text-sm"
                    type="text"
                    value={confirm}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      setConfirm(e?.currentTarget?.value)
                    }}
                  />
                </label>
              </div>

              <div className="mt-6 flex gap-3 justify-end">
                <Button className="text-sm" onClick={handleClose} type="button">
                  Cancel
                </Button>
                <Button
                  className="text-sm flex items-center gap-2 font-medium"
                  disabled={
                    confirm !== connectionName || isLoading || isKickedOff
                  }
                  variant="danger"
                  onClick={() => {
                    setIsKickedOff(true)
                    execute({
                      connectionId: connection?.id,
                      orgId: org?.id,
                      path,
                    })
                  }}
                >
                  {isLoading ? (
                    <>
                      <SpinnerSVG /> Disconnecting GitHub...
                    </>
                  ) : (
                    <>
                      <XCircleIcon size="16" /> Disconnect GitHub
                    </>
                  )}
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
    </>
  )
}
