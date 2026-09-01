export default {
  title: 'Branches/BranchInputs',
}

import { BranchInputs } from './BranchInputs'

const appConfig = {
  input: {
    input_groups: [
      {
        id: 'general',
        display_name: 'General settings',
        description: 'Core application configuration',
        input_ids: ['app-name', 'region'],
      },
      {
        id: 'database',
        display_name: 'Database settings',
        description: 'Database connection configuration',
        input_ids: ['database-password'],
      },
    ],
    inputs: [
      {
        id: 'app-name',
        name: 'app_name',
        display_name: 'Application name',
        description: 'The name of the application',
        default: 'example-app',
        required: true,
        sensitive: false,
        source: 'vendor',
        group_id: 'general',
      },
      {
        id: 'region',
        name: 'region',
        display_name: 'AWS region',
        description: 'The AWS region to deploy to',
        default: 'us-east-1',
        required: true,
        sensitive: false,
        source: 'installer',
        group_id: 'general',
      },
      {
        id: 'database-password',
        name: 'database_password',
        display_name: 'Database password',
        description: 'The database password',
        default: '',
        required: true,
        sensitive: true,
        source: 'installer',
        group_id: 'database',
      },
    ],
  },
} as any

export const Default = () => <BranchInputs appConfig={appConfig} />

export const Loading = () => <BranchInputs isLoading />

export const Empty = () => <BranchInputs />

export const Error = () => <BranchInputs isError />
