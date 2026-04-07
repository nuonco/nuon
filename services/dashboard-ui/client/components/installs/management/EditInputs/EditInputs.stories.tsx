export default {
  title: 'Installs/EditInputs',
}

import { useRef } from 'react'
import { ModalStory } from '@/components/__stories__/helpers'
import { EditInputsFormModal } from './EditInputs'

const noop = () => {}

const mockInstall = {
  id: 'inst-123',
  name: 'prod-acme',
  app_id: 'app-123',
  app_config_id: 'cfg-123',
} as any

export const Loading = () => {
  const formRef = useRef<HTMLFormElement>(null)
  const clearDraftRef = useRef<(() => void) | null>(null)

  return (
    <ModalStory>
      <EditInputsFormModal
        install={mockInstall}
        config={undefined}
        isLoading={true}
        error={null}
        isSubmitting={false}
        actionError={null}
        onFormSubmit={noop}
        onClose={noop}
        formRef={formRef}
        clearDraftRef={clearDraftRef}
        selectedRole=""
        onRoleChange={noop}
        deployDependents={true}
        onDeployDependentsChange={noop}
        onMutate={noop as any}
      />
    </ModalStory>
  )
}

export const WithError = () => {
  const formRef = useRef<HTMLFormElement>(null)
  const clearDraftRef = useRef<(() => void) | null>(null)

  return (
    <ModalStory>
      <EditInputsFormModal
        install={mockInstall}
        config={undefined}
        isLoading={false}
        error={{ error: 'Unable to load app configuration' }}
        isSubmitting={false}
        actionError={null}
        onFormSubmit={noop}
        onClose={noop}
        formRef={formRef}
        clearDraftRef={clearDraftRef}
        selectedRole=""
        onRoleChange={noop}
        deployDependents={true}
        onDeployDependentsChange={noop}
        onMutate={noop as any}
      />
    </ModalStory>
  )
}
