export default {
  title: 'Apps/Config/AppInputs',
}

import { AppInputs } from './AppInputs'

const mockAppConfig = {
  input: {
    input_groups: [
      {
        id: 'group-general',
        display_name: 'General settings',
        description: 'Core application configuration',
        input_ids: ['input-1', 'input-2'],
      },
      {
        id: 'group-database',
        display_name: 'Database settings',
        description: 'Database connection configuration',
        input_ids: ['input-3'],
      },
    ],
    inputs: [
      {
        id: 'input-1',
        name: 'app_name',
        display_name: 'Application name',
        description: 'The name of the application',
        default: 'my-app',
        required: true,
        sensitive: false,
        source: 'vendor',
        group_id: 'group-general',
      },
      {
        id: 'input-2',
        name: 'region',
        display_name: 'AWS region',
        description: 'The AWS region to deploy to',
        default: 'us-east-1',
        required: true,
        sensitive: false,
        source: 'installer',
        group_id: 'group-general',
      },
      {
        id: 'input-3',
        name: 'db_password',
        display_name: 'Database password',
        description: 'Password for the database connection',
        default: '',
        required: true,
        sensitive: true,
        source: 'installer',
        group_id: 'group-database',
      },
    ],
  },
} as any

export const Default = () => <AppInputs appConfig={mockAppConfig} />
