import { ModalStory } from '@/components/__stories__/helpers'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAppInputConfig } from '@/types'
import { InstallForm } from './InstallForm'
import { useInstallForm } from './useInstallForm'
import type { InstallPlatform } from './schema'

export default { title: 'Installs/InstallForm' }

const inputConfig = {
  id: 'cfg-1',
  input_groups: [
    {
      id: 'g1',
      display_name: 'Application settings',
      description: 'Configure the application for this install.',
      index: 0,
      app_inputs: [
        {
          id: 'i1',
          name: 'database_url',
          display_name: 'Database URL',
          description: 'Connection string for the primary database.',
          type: 'string',
          required: true,
          source: 'vendor',
          index: 0,
        },
        {
          id: 'i2',
          name: 'enable_backups',
          display_name: 'Enable backups',
          type: 'bool',
          default: 'true',
          source: 'vendor',
          index: 1,
        },
        {
          id: 'i3',
          name: 'extra_config',
          display_name: 'Extra config',
          description: 'Optional YAML overrides.',
          type: 'yaml',
          required: false,
          source: 'vendor',
          index: 2,
        },
      ],
    },
  ],
} as unknown as TAppInputConfig

const CreateStory = ({
  platform,
  withInputs,
}: {
  platform: InstallPlatform
  withInputs?: boolean
}) => {
  const { form } = useInstallForm({
    mode: 'create',
    platform,
    inputConfig: withInputs ? inputConfig : undefined,
    onSubmit: () => {},
  })

  return (
    <div className="p-6">
      <InstallForm
        form={form}
        mode="create"
        platform={platform}
        inputConfig={withInputs ? inputConfig : undefined}
      />
    </div>
  )
}

export const CreateAws = () => <CreateStory platform="aws" />
export const CreateAwsWithInputs = () => (
  <CreateStory platform="aws" withInputs />
)
export const CreateAzure = () => <CreateStory platform="azure" />
export const CreateGcp = () => <CreateStory platform="gcp" />

const CreateModalHarness = (props: IModal) => {
  const { form, canSubmit } = useInstallForm({
    mode: 'create',
    platform: 'gcp',
    onSubmit: () => {},
  })

  return (
    <Modal
      heading="Create install"
      primaryActionTrigger={{
        children: 'Create install',
        disabled: !canSubmit,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <InstallForm form={form} mode="create" platform="gcp" />
    </Modal>
  )
}

export const CreateModal = () => (
  <ModalStory>
    <CreateModalHarness />
  </ModalStory>
)
