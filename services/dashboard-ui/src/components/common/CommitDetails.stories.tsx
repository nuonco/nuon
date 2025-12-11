/* eslint-disable react/no-unescaped-entities */
import { CommitDetails } from './CommitDetails'
import { LabeledValue } from './LabeledValue'
import { Text } from './Text'

const mockCommit = {
  id: 'commit_1234567890abcdef',
  sha: 'a1b2c3d4e5f6',
  author_name: 'Jane Smith',
  author_email: 'jane.smith@example.com',
  message: 'Fix authentication bug and update user permissions',
  created_at: '2024-03-15T14:30:00Z',
  vcs_connection_id: 'vcs_conn_123'
}

const longCommitMessage = {
  ...mockCommit,
  sha: 'x9y8z7w6v5u4',
  message: 'Implement comprehensive error handling system with detailed logging, user-friendly error messages, retry mechanisms, and proper error boundary components for React applications'
}

const shortCommit = {
  ...mockCommit,
  sha: 'f1e2d3c4b5a6',
  author_name: 'Bob Wilson',
  author_email: 'bob@dev.co',
  message: 'Fix typo'
}

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic CommitDetails Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        CommitDetails displays git commit information in a tooltip when you hover
        over the commit SHA. It shows the author, commit message, and date in a
        clean, accessible format. Perfect for build headers, deploy information,
        or any interface showing git commit data.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Standard Commit Display</h4>
      <div className="space-y-3 max-w-md">
        <LabeledValue label="Commit SHA">
          <CommitDetails commit={mockCommit} />
        </LabeledValue>
        
        <LabeledValue label="Latest Commit">
          <CommitDetails commit={shortCommit} />
        </LabeledValue>
      </div>
      <Text variant="subtext" theme="neutral">
        Hover over the commit SHA to see detailed information
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Truncated SHA display (first 6 characters) with full details in tooltip</li>
        <li>Author name and email in tooltip</li>
        <li>Commit message with automatic truncation for long messages</li>
        <li>Formatted timestamp showing when the commit was created</li>
        <li>Graceful handling of missing or null commit data</li>
      </ul>
    </div>
  </div>
)

export const MessageHandling = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Commit Message Handling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        CommitDetails automatically handles different commit message lengths.
        Long commit messages are truncated in the tooltip to maintain a clean
        appearance while still being readable.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Different Message Lengths</h4>
      <div className="space-y-4 max-w-md">
        <LabeledValue label="Short Message">
          <CommitDetails commit={shortCommit} />
        </LabeledValue>
        
        <LabeledValue label="Standard Message">
          <CommitDetails commit={mockCommit} />
        </LabeledValue>
        
        <LabeledValue label="Long Message">
          <CommitDetails commit={longCommitMessage} />
        </LabeledValue>
      </div>
      <Text variant="subtext" theme="neutral">
        Messages longer than 180px width are automatically truncated in the tooltip
      </Text>
    </div>
  </div>
)

export const IntegrationExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Integration Examples</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        CommitDetails is commonly used in build headers, deployment summaries,
        and anywhere git commit information needs to be displayed. It integrates
        well with other components and layouts.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Build Header Style</h4>
      <div className="p-4 border rounded-lg">
        <Text variant="h3" weight="stronger" className="mb-4">
          Component Build #142
        </Text>
        <div className="flex gap-6">
          <LabeledValue label="Status">
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 bg-green-500 rounded-full"></div>
              <Text theme="success">Success</Text>
            </div>
          </LabeledValue>
          
          <LabeledValue label="Duration">
            <Text>2m 34s</Text>
          </LabeledValue>
          
          <LabeledValue label="Commit SHA">
            <CommitDetails commit={mockCommit} />
          </LabeledValue>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Deployment Summary</h4>
      <div className="p-4 border rounded-lg">
        <Text variant="h3" weight="stronger" className="mb-4">
          Production Deployment
        </Text>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <LabeledValue label="Environment">
            <Text>production</Text>
          </LabeledValue>
          
          <LabeledValue label="Deployed At">
            <Text variant="subtext">2024-03-15 at 2:30 PM UTC</Text>
          </LabeledValue>
          
          <LabeledValue label="Source Commit">
            <CommitDetails commit={mockCommit} />
          </LabeledValue>
          
          <LabeledValue label="Deploy ID">
            <Text family="mono">deploy_abc123</Text>
          </LabeledValue>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Multiple Commits</h4>
      <div className="p-4 border rounded-lg">
        <Text variant="h3" weight="stronger" className="mb-4">
          Release v2.1.0
        </Text>
        <div className="space-y-3">
          <LabeledValue label="Head Commit">
            <CommitDetails commit={mockCommit} />
          </LabeledValue>
          
          <LabeledValue label="Previous Release">
            <CommitDetails commit={shortCommit} />
          </LabeledValue>
          
          <LabeledValue label="Feature Branch">
            <CommitDetails commit={longCommitMessage} />
          </LabeledValue>
        </div>
      </div>
    </div>
  </div>
)

export const EdgeCases = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Edge Cases & Error Handling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        CommitDetails gracefully handles missing data, null values, and edge
        cases. When commit data is unavailable, the component simply doesn't
        render anything.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Missing Data Scenarios</h4>
      <div className="space-y-4 max-w-md">
        <LabeledValue label="No Commit Data">
          <CommitDetails commit={null} />
          <Text variant="subtext" theme="neutral">
            (Component renders nothing when commit is null)
          </Text>
        </LabeledValue>
        
        <LabeledValue label="Valid Commit">
          <CommitDetails commit={mockCommit} />
        </LabeledValue>
        
        <LabeledValue label="Undefined Commit">
          <CommitDetails commit={undefined} />
          <Text variant="subtext" theme="neutral">
            (Component renders nothing when commit is undefined)
          </Text>
        </LabeledValue>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Always wrap CommitDetails in a LabeledValue for proper context</li>
        <li>Handle null/undefined commit data at the parent level if needed</li>
        <li>Use in contexts where users understand git concepts</li>
        <li>Consider the tooltip interaction pattern for touch devices</li>
        <li>Provide fallback UI when commit information is critical</li>
      </ul>
    </div>
  </div>
)