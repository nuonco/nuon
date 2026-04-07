export default {
  title: 'Installs/Forms/InputConfigFields',
}

import { InputConfigFields } from './InputConfigFields'
import type { TAppInputConfig } from '@/types'

const mockInputConfig: TAppInputConfig = {
  id: 'config-1',
  input_groups: [
    {
      id: 'group-1',
      display_name: 'Database settings',
      description: 'Configure your database connection',
      index: 0,
      app_inputs: [
        {
          id: 'input-1',
          name: 'db_host',
          display_name: 'Database host',
          description: 'The hostname or IP of your database server',
          type: 'string',
          required: true,
          default: 'localhost',
          index: 0,
          source: 'vendor',
        },
        {
          id: 'input-2',
          name: 'db_port',
          display_name: 'Database port',
          description: 'The port number',
          type: 'number',
          required: false,
          default: '5432',
          index: 1,
          source: 'vendor',
        },
        {
          id: 'input-3',
          name: 'db_ssl',
          display_name: 'Enable SSL',
          type: 'bool',
          required: false,
          default: 'true',
          index: 2,
          source: 'vendor',
        },
      ],
    },
  ],
} as TAppInputConfig

export const Default = () => (
  <form className="max-w-2xl p-6 flex flex-col gap-6">
    <InputConfigFields inputConfig={mockInputConfig} />
  </form>
)

export const WithDraftValues = () => (
  <form className="max-w-2xl p-6 flex flex-col gap-6">
    <InputConfigFields
      inputConfig={mockInputConfig}
      draftValues={{ 'inputs:db_host': 'prod-db.example.com', 'inputs:db_port': '5433' }}
    />
  </form>
)
