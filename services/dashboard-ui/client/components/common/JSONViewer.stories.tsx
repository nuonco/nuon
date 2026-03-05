import { JSONViewer } from './JSONViewer'
import { Text } from './Text'

const sampleData = {
  name: 'John Doe',
  age: 30,
  email: 'john.doe@example.com',
  isActive: true,
  roles: ['admin', 'user', 'developer'],
  profile: {
    avatar: 'https://example.com/avatar.jpg',
    bio: 'Software engineer with 10+ years of experience',
    preferences: {
      theme: 'dark',
      language: 'en',
      notifications: {
        email: true,
        push: false,
        sms: true
      }
    }
  },
  projects: [
    {
      id: 1,
      name: 'Dashboard UI',
      status: 'active',
      technologies: ['React', 'TypeScript', 'Tailwind'],
      metrics: {
        lines_of_code: 15420,
        test_coverage: 0.87,
        performance_score: 95
      }
    },
    {
      id: 2,
      name: 'API Gateway',
      status: 'maintenance',
      technologies: ['Node.js', 'Express', 'PostgreSQL'],
      metrics: {
        lines_of_code: 8340,
        test_coverage: 0.93,
        performance_score: 88
      }
    }
  ],
  metadata: {
    created_at: '2024-01-15T10:30:45.123Z',
    updated_at: '2024-03-20T14:22:33.456Z',
    version: '1.2.3'
  }
}

const configData = {
  database: {
    host: 'localhost',
    port: 5432,
    name: 'dashboard_db',
    ssl: true,
    pool: {
      min: 2,
      max: 10,
      idle_timeout: 30000
    },
    credentials: {
      username: 'db_user',
      password: '••••••••',
      connection_string: 'postgresql://user:pass@localhost:5432/db'
    }
  },
  cache: {
    redis: {
      host: 'redis.example.com',
      port: 6379,
      ttl: 3600,
      cluster_mode: false
    }
  },
  features: {
    analytics: true,
    monitoring: true,
    debug_mode: false,
    feature_flags: ['new_ui', 'advanced_search', 'beta_features']
  }
}

const apiResponse = {
  success: true,
  data: {
    users: [
      { id: 1, name: 'Alice Johnson', role: 'admin', last_login: '2024-03-20T09:15:00Z' },
      { id: 2, name: 'Bob Smith', role: 'user', last_login: '2024-03-19T16:42:00Z' },
      { id: 3, name: 'Carol Davis', role: 'moderator', last_login: '2024-03-20T11:30:00Z' }
    ],
    pagination: {
      page: 1,
      per_page: 3,
      total: 25,
      total_pages: 9
    },
    filters: {
      active: true,
      roles: ['admin', 'user', 'moderator'],
      date_range: {
        start: '2024-03-01',
        end: '2024-03-20'
      }
    }
  },
  meta: {
    request_id: 'req_abc123def456',
    timestamp: '2024-03-20T14:30:00.789Z',
    execution_time: 0.142
  }
}

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic JSONViewer Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        JSONViewer provides an interactive way to explore complex JSON data structures.
        Unlike static code blocks, users can collapse and expand nodes, making it perfect
        for debugging, API responses, and configuration management.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">User Profile Data</h4>
      <div className="max-w-4xl">
        <JSONViewer data={sampleData} />
      </div>
      <Text variant="subtext" theme="neutral">
        Click on object keys to expand/collapse nested structures
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Interactive expand/collapse functionality</li>
        <li>Automatic dark/light theme detection</li>
        <li>Data type indicators for better understanding</li>
        <li>Object size information for arrays and objects</li>
        <li>Copy to clipboard functionality</li>
        <li>Customizable expansion levels</li>
      </ul>
    </div>
  </div>
)

export const ConfigurationData = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Configuration Management</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        JSONViewer excels at displaying complex configuration objects with nested
        structures. The interactive nature makes it easy to navigate through
        database settings, feature flags, and service configurations.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Application Configuration</h4>
      <div className="max-w-4xl">
        <JSONViewer data={configData} expanded={1} />
      </div>
      <Text variant="subtext" theme="neutral">
        Configuration starts collapsed at level 1 for better overview
      </Text>
    </div>
  </div>
)

export const APIResponses = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">API Response Exploration</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Perfect for debugging API responses and understanding complex data structures
        returned from backend services. The interactive format makes it easy to
        drill down into specific data points.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">User List API Response</h4>
      <div className="max-w-4xl">
        <JSONViewer 
          data={apiResponse} 
          showSize={true}
          showToolbar={true}
        />
      </div>
      <Text variant="subtext" theme="neutral">
        API response with sorted keys and object size indicators
      </Text>
    </div>
  </div>
)

export const CustomizationOptions = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Customization Options</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        JSONViewer offers various customization options to control how data is
        displayed, including expansion levels, data type visibility, clipboard
        functionality, and key formatting.
      </p>
    </div>

    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div className="space-y-4">
        <h4 className="text-sm font-medium">Minimal Configuration</h4>
        <JSONViewer 
          data={sampleData.profile}
          showDataTypes={false}
          showSize={false}
          showCopy={false}
          className="max-w-full"
        />
        <Text variant="subtext" theme="neutral">
          Hidden data types, object sizes, and clipboard with quoted keys
        </Text>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Fully Collapsed</h4>
        <JSONViewer 
          data={sampleData.projects}
          expanded={false}
          className="max-w-full"
        />
        <Text variant="subtext" theme="neutral">
          All nodes start in collapsed state
        </Text>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Deep Expansion</h4>
      <div className="max-w-4xl">
        <JSONViewer 
          data={configData.database}
          expanded={10}
          showToolbar={true}
        />
      </div>
      <Text variant="subtext" theme="neutral">
        All nested levels expanded with sorted and quoted keys
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Customization Props:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li><code>expanded</code>: Control how many levels are expanded by default (number) or fully expand/collapse (boolean)</li>
        <li><code>indent</code>: Number of spaces for indentation (default: 2)</li>
        <li><code>showDataTypes</code>: Show/hide data type indicators</li>
        <li><code>showSize</code>: Show/hide object and array sizes</li>
        <li><code>showCopy</code>: Enable/disable copy to clipboard</li>
        <li><code>showToolbar</code>: Enable interactive toolbar with expand/collapse controls</li>
        <li><code>expandIconType</code>: Icon style for expand/collapse (&lsquo;square&rsquo;, &lsquo;circle&rsquo;, &lsquo;arrow&rsquo;)</li>
      </ul>
    </div>
  </div>
)

