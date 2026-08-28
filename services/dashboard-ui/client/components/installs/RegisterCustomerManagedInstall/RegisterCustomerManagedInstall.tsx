import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Modal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import {
  registerCustomerManagedInstall,
  type TInstallationRegistration,
} from '@/lib'
import type { TAPIError } from '@/types'

const RegistrationModal = ({ modalId }: { modalId?: string }) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const [registration, setRegistration] = useState<TInstallationRegistration>()
  const [fileName, setFileName] = useState('')
  const [fileError, setFileError] = useState('')
  const mutation = useMutation({
    mutationFn: (body: TInstallationRegistration) =>
      registerCustomerManagedInstall({ body, orgId: org.id }),
    onSuccess: ({ install }) => {
      queryClient.invalidateQueries({ queryKey: ['installs'] })
      addToast(
        <Toast heading="Customer-managed install registered" theme="success">
          <Link href={`/${org.id}/installs/${install.id}`}>{install.name}</Link>
        </Toast>
      )
      removeModal(modalId)
    },
  })

  const selectFile = async (file?: File) => {
    setRegistration(undefined)
    setFileError('')
    setFileName(file?.name ?? '')
    if (!file) return
    try {
      const parsed = JSON.parse(await file.text()) as TInstallationRegistration
      if (
        !parsed.registration_id ||
        !parsed.archive_digest ||
        !parsed.install_id
      ) {
        throw new Error('The file is not a Nuon installation registration.')
      }
      setRegistration(parsed)
    } catch (error) {
      setFileError(
        error instanceof Error
          ? error.message
          : 'Unable to read registration file.'
      )
    }
  }

  return (
    <Modal
      modalId={modalId}
      heading="Register customer-managed install"
      primaryActionTrigger={{
        children: mutation.isPending ? 'Registering…' : 'Register install',
        disabled: !registration || mutation.isPending,
        onClick: () => registration && mutation.mutate(registration),
        variant: 'primary',
      }}
    >
      <div className="flex flex-col gap-4 mb-6">
        <Text theme="neutral">
          Upload the installation registration downloaded from the customer
          portal after installation completed.
        </Text>
        <label className="flex cursor-pointer flex-col items-center gap-2 rounded border border-dashed p-6 text-center">
          <Icon variant="CloudArrowUpIcon" size={24} />
          <Text weight="strong">Choose installation-registration.json</Text>
          {fileName ? <Text variant="subtext">{fileName}</Text> : null}
          <input
            className="sr-only"
            type="file"
            accept="application/json,.json"
            onChange={(event) => selectFile(event.target.files?.[0])}
          />
        </label>
        {fileError ? <Banner theme="error">{fileError}</Banner> : null}
        {mutation.error ? (
          <Banner theme="error">
            {(mutation.error as TAPIError).description ||
              (mutation.error as TAPIError).error ||
              'Unable to register this installation.'}
          </Banner>
        ) : null}
        {registration ? (
          <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 rounded border p-4">
            <Text theme="neutral">Deployment</Text>
            <Badge variant="code">{registration.deployment_id}</Badge>
            <Text theme="neutral">Install ID</Text>
            <Text>{registration.install_id}</Text>
            <Text theme="neutral">Cloud</Text>
            <Text>
              {registration.cloud.provider} · {registration.cloud.account_id} ·{' '}
              {registration.cloud.region}
            </Text>
            <Text theme="neutral">Stack</Text>
            <Text>{registration.stack.name}</Text>
            <Text theme="neutral">Bundle archive</Text>
            <Text className="break-all" variant="subtext">
              {registration.archive_digest}
            </Text>
          </div>
        ) : null}
      </div>
    </Modal>
  )
}

export const RegisterCustomerManagedInstallButton = (
  props: IButtonAsButton
) => {
  const { addModal } = useSurfaces()
  return (
    <Button onClick={() => addModal(<RegistrationModal />)} {...props}>
      <Icon variant="CloudArrowUpIcon" size={16} />
      Register offline install
    </Button>
  )
}
