export default {
  title: 'Installs/EditInputs',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { OrgContext } from '@/providers/org-provider'
import type { TAppInputConfig, TInstall, TOrg } from '@/types'
import { EditInstallModal } from './EditInputs'

const mockOrg = {
  org: { id: 'orgstory', name: 'Story org' } as TOrg,
  refresh: () => {},
}

const mockInstall = {
  id: 'inst-123',
  name: 'prod-acme',
  app_id: 'app-123',
  app_config_id: 'cfg-123',
} as unknown as TInstall

const inputConfig = {
  id: 'cfg-123',
  input_groups: [
    {
      id: 'group-app',
      display_name: 'Application settings',
      description: 'Core application configuration',
      index: 0,
      app_inputs: [
        {
          id: 'app-name',
          name: 'app_name',
          display_name: 'App name',
          description: 'Display name for the application',
          type: 'string',
          required: true,
          default: 'acme-app',
          index: 0,
          source: 'vendor',
        },
        {
          id: 'replica-count',
          name: 'replica_count',
          display_name: 'Replica count',
          type: 'number',
          required: false,
          default: '2',
          index: 1,
          source: 'vendor',
        },
        {
          id: 'enable-metrics',
          name: 'enable_metrics',
          display_name: 'Enable metrics',
          type: 'bool',
          required: false,
          default: 'true',
          index: 2,
          source: 'vendor',
        },
      ],
    },
    {
      id: 'group-net',
      display_name: 'Networking',
      description: 'Ingress and network configuration',
      index: 1,
      app_inputs: [
        {
          id: 'extra-config',
          name: 'extra_config',
          display_name: 'Extra config',
          description: 'Optional YAML overrides.',
          type: 'yaml',
          required: false,
          index: 0,
          source: 'vendor',
        },
      ],
    },
  ],
} as unknown as TAppInputConfig

const noop = async () => {}

export const WithInputs = () => (
  <OrgContext.Provider value={mockOrg}>
    <ModalStory>
      <EditInstallModal
        install={mockInstall}
        inputConfig={inputConfig}
        isSubmitting={false}
        submitError={null}
        onSubmitName={noop}
        onSubmitInputs={noop}
      />
    </ModalStory>
  </OrgContext.Provider>
)

export const WithNameField = () => (
  <OrgContext.Provider value={mockOrg}>
    <ModalStory>
      <EditInstallModal
        install={mockInstall}
        inputConfig={inputConfig}
        showNameField
        isSubmitting={false}
        submitError={null}
        onSubmitName={noop}
        onSubmitInputs={noop}
      />
    </ModalStory>
  </OrgContext.Provider>
)

export const WithError = () => (
  <OrgContext.Provider value={mockOrg}>
    <ModalStory>
      <EditInstallModal
        install={mockInstall}
        inputConfig={inputConfig}
        isSubmitting={false}
        submitError={{ error: 'Unable to update install' } as any}
        onSubmitName={noop}
        onSubmitInputs={noop}
      />
    </ModalStory>
  </OrgContext.Provider>
)
