export default {
  title: 'Common/PropertyGrid',
}

import { PropertyGrid } from './PropertyGrid'
import { Badge } from './Badge'
import { Link } from './Link'
import { Text } from './Text'

export const Default = () => (
  <PropertyGrid 
    values={[
      { key: 'name', value: 'example-app' },
      { key: 'version', value: '1.0.0' },
      { key: 'status', value: 'active' }
    ]} 
  />
)

export const SimpleProperties = () => (
  <PropertyGrid 
    values={[
      { name: 'Version', value: '1.0.0', type: 'string' },
      { name: 'Status', value: 'Active', type: 'string' },
      { name: 'Created', value: '2024-01-09', type: 'date' },
      { name: 'Port', value: '8080', type: 'number' },
      { name: 'Enabled', value: 'true', type: 'boolean' }
    ]} 
  />
)

export const CustomColumns = () => {
  const appInputs = [
    {
      name: 'DATABASE_URL',
      defaultValue: 'postgresql://localhost:5432/mydb',
      required: true,
      sensitive: false,
      description: 'Database connection string'
    },
    {
      name: 'API_KEY',
      defaultValue: '',
      required: true,
      sensitive: true,
      description: 'Third-party API authentication key'
    },
    {
      name: 'DEBUG_MODE',
      defaultValue: 'false',
      required: false,
      sensitive: false,
      description: 'Enable debug logging'
    }
  ]

  return (
    <PropertyGrid
      values={appInputs}
      columns={[
        { key: 'name', header: 'Input Name' },
        { key: 'defaultValue', header: 'Default Value' },
        { 
          key: 'required', 
          header: 'Required',
          render: (value) => value ? 'Yes' : 'No'
        },
        { 
          key: 'sensitive', 
          header: 'Sensitive',
          render: (value) => value ? 'Yes' : 'No'
        },
        { key: 'description', header: 'Description' }
      ]}
    />
  )
}

export const ComplexData = () => {
  const complexData = [
    {
      name: 'Status',
      value: <Badge theme="success">Active</Badge>,
      updated: '2 minutes ago'
    },
    {
      name: 'Documentation',
      value: <Link href="https://docs.nuon.co" target="_blank">View docs</Link>,
      updated: '1 hour ago'
    },
    {
      name: 'Environment',
      value: <Badge theme="info">Production</Badge>,
      updated: '5 minutes ago'
    },
    {
      name: 'Health Check',
      value: <Text theme="success" weight="strong">Healthy</Text>,
      updated: 'Just now'
    }
  ]

  return (
    <PropertyGrid
      values={complexData}
      columns={[
        { key: 'name', header: 'Property' },
        { key: 'value', header: 'Current Value' },
        { key: 'updated', header: 'Last Updated' }
      ]}
    />
  )
}

export const AppInputsExample = () => {
  const appInputsData = [
    {
      name: 'POSTGRES_HOST',
      default: 'localhost',
      required: <Badge theme="success">Yes</Badge>,
      sensitive: <Badge theme="neutral">No</Badge>,
      description: 'PostgreSQL database host'
    },
    {
      name: 'POSTGRES_PASSWORD',
      default: '',
      required: <Badge theme="success">Yes</Badge>,
      sensitive: <Badge theme="error">Yes</Badge>,
      description: 'PostgreSQL database password'
    },
    {
      name: 'LOG_LEVEL',
      default: 'info',
      required: <Badge theme="neutral">No</Badge>,
      sensitive: <Badge theme="neutral">No</Badge>,
      description: 'Application logging level'
    },
    {
      name: 'FEATURE_FLAGS',
      default: '{}',
      required: <Badge theme="neutral">No</Badge>,
      sensitive: <Badge theme="neutral">No</Badge>,
      description: 'JSON object with feature flag configuration'
    }
  ]

  return (
    <PropertyGrid
      values={appInputsData}
      columns={[
        { key: 'name', header: 'Input Name' },
        { key: 'default', header: 'Default' },
        { key: 'required', header: 'Required' },
        { key: 'sensitive', header: 'Sensitive' },
        { key: 'description', header: 'Description' }
      ]}
      gridTemplate="minmax(150px, 2fr) minmax(120px, 1.5fr) minmax(80px, max-content) minmax(80px, max-content) minmax(200px, 3fr)"
    />
  )
}

export const WithLongContent = () => {
  const longContentData = [
    {
      property: 'Docker Image',
      value: 'registry.example.com/myapp:v1.2.3-production-build-abc123def456',
      category: 'Runtime'
    },
    {
      property: 'Environment Variables',
      value: 'NODE_ENV=production, DEBUG=false, API_URL=https://api.example.com/v1, DATABASE_URL=postgresql://user:pass@host:5432/db',
      category: 'Configuration'
    },
    {
      property: 'Command',
      value: 'npm start --production --max-old-space-size=2048 --optimize-for-size',
      category: 'Runtime'
    }
  ]

  return (
    <PropertyGrid
      values={longContentData}
      columns={[
        { key: 'property', header: 'Property', className: 'min-w-32' },
        { key: 'value', header: 'Value', className: 'font-mono text-sm' },
        { key: 'category', header: 'Category' }
      ]}
    />
  )
}

export const Empty = () => (
  <PropertyGrid 
    values={[]} 
    emptyStateProps={{
      variant: 'table',
      emptyTitle: 'No properties',
      emptyMessage: 'No property data to display'
    }}
  />
)

export const AutoDetectedColumns = () => (
  <PropertyGrid 
    values={[
      { applicationName: 'My App', deploymentStatus: 'Running', lastDeploy: '2024-01-09' },
      { applicationName: 'API Service', deploymentStatus: 'Stopped', lastDeploy: '2024-01-08' },
      { applicationName: 'Worker', deploymentStatus: 'Running', lastDeploy: '2024-01-09' }
    ]} 
  />
)

export const CustomGridTemplate = () => (
  <PropertyGrid 
    values={[
      { applicationName: 'My App', deploymentStatus: 'Running', lastDeploy: '2024-01-09' },
      { applicationName: 'API Service', deploymentStatus: 'Stopped', lastDeploy: '2024-01-08' },
      { applicationName: 'Worker', deploymentStatus: 'Running', lastDeploy: '2024-01-09' }
    ]}
    gridTemplate="auto auto 1fr"  // Application Name gets 2 parts, others get 1 part each
  />
)

export const BalancedColumns = () => (
  <PropertyGrid 
    values={[
      { applicationName: 'My App', deploymentStatus: 'Running', lastDeploy: '2024-01-09' },
      { applicationName: 'API Service', deploymentStatus: 'Stopped', lastDeploy: '2024-01-08' },
      { applicationName: 'Long Application Name Here', deploymentStatus: 'Running', lastDeploy: '2024-01-09' }
    ]}
    gridTemplate="minmax(150px, 2fr) minmax(100px, 1fr) minmax(100px, 1fr)"  // More control over min/max widths
  />
)
