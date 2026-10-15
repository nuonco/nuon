import { useState } from 'react'
import { ModalStory } from '@/components/__stories__/helpers'
import type { IModal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'
import {
  CollectorCredentialModal,
  type TCollectorCredentialValues,
} from './CollectorCredential'

export default { title: 'Features / Installs / Collector credential' }

const CredentialModal = ({
  issued = false,
  ...props
}: IModal & { issued?: boolean }) => {
  const { removeModal } = useSurfaces()
  const [credential, setCredential] = useState<
    (TCollectorCredentialValues & { token: string }) | null
  >(
    issued
      ? {
          duration: '8760h',
          token: 'example-not-a-real-token',
        }
      : null
  )
  return (
    <CollectorCredentialModal
      installId="install-acme"
      isPending={false}
      error={null}
      credential={credential}
      {...props}
      onSubmit={(values) =>
        setCredential({ ...values, token: 'example-not-a-real-token' })
      }
      onDone={() => removeModal(props.modalId)}
    />
  )
}

export const Default = () => (
  <ModalStory>
    <CredentialModal />
  </ModalStory>
)
export const Created = () => (
  <ModalStory>
    <CredentialModal issued />
  </ModalStory>
)
