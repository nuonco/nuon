/* eslint-disable react/no-unescaped-entities */
import { Pagination } from './Pagination'
import { Text } from './Text'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Pagination Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Pagination provides navigation controls for paged data sets. It
        automatically manages URL parameters to maintain pagination state
        across page navigation and browser refreshes. The component includes
        loading states and automatic URL updates.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Default Pagination</h4>
      <div className="flex justify-center p-6 border rounded-lg">
        <Pagination />
      </div>
      <Text variant="subtext" theme="neutral" className="text-center">
        Default state - Previous disabled, Next disabled (no data)
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Default Behavior:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Previous button disabled when at first page (offset = 0)</li>
        <li>Next button disabled when hasNext is false</li>
        <li>Automatically updates URL parameters for state persistence</li>
        <li>Loading states prevent multiple rapid clicks</li>
      </ul>
    </div>
  </div>
)

export const PaginationStates = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Pagination States</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Pagination components show different states based on the current
        position in the data set. The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          hasNext
        </code>{' '}
        and{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          offset
        </code>{' '}
        props control which buttons are enabled.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">First Page (offset = 0, hasNext = true)</h4>
      <div className="flex justify-center p-6 border rounded-lg">
        <Pagination offset={0} hasNext={true} />
      </div>
      <Text variant="subtext" theme="neutral" className="text-center">
        Previous disabled, Next enabled
      </Text>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Middle Page (offset = 10, hasNext = true)</h4>
      <div className="flex justify-center p-6 border rounded-lg">
        <Pagination offset={10} hasNext={true} />
      </div>
      <Text variant="subtext" theme="neutral" className="text-center">
        Both Previous and Next enabled
      </Text>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Last Page (offset = 20, hasNext = false)</h4>
      <div className="flex justify-center p-6 border rounded-lg">
        <Pagination offset={20} hasNext={false} />
      </div>
      <Text variant="subtext" theme="neutral" className="text-center">
        Previous enabled, Next disabled
      </Text>
    </div>
  </div>
)

export const Positioning = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Pagination Positioning</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          position
        </code>{' '}
        prop controls the alignment of pagination controls within their
        container. This is useful for matching pagination placement with table
        layouts and content alignment.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Left Aligned</h4>
      <div className="flex p-6 border rounded-lg">
        <Pagination offset={10} hasNext={true} position="left" />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Center Aligned (Default)</h4>
      <div className="flex p-6 border rounded-lg">
        <Pagination offset={10} hasNext={true} position="center" />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Right Aligned</h4>
      <div className="flex p-6 border rounded-lg">
        <Pagination offset={10} hasNext={true} position="right" />
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm mt-6">
      <div>
        <strong>left:</strong> Aligns pagination controls to the left of the
        container
      </div>
      <div>
        <strong>center:</strong> Centers pagination controls in the container
        (default)
      </div>
      <div>
        <strong>right:</strong> Aligns pagination controls to the right of the
        container
      </div>
    </div>
  </div>
)

export const CustomConfiguration = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Pagination Configuration</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Pagination can be customized with different page sizes and URL parameter
        names. The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          limit
        </code>{' '}
        prop controls page size, while{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          param
        </code>{' '}
        allows custom URL parameter names.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Different Page Sizes</h4>
      <div className="space-y-4">
        <div className="p-4 border rounded-lg">
          <div className="flex justify-between items-center mb-3">
            <Text variant="subtext">10 items per page (default)</Text>
            <Pagination offset={0} hasNext={true} limit={10} />
          </div>
          <Text variant="label" theme="neutral">
            Standard page size for most lists
          </Text>
        </div>

        <div className="p-4 border rounded-lg">
          <div className="flex justify-between items-center mb-3">
            <Text variant="subtext">25 items per page</Text>
            <Pagination offset={0} hasNext={true} limit={25} />
          </div>
          <Text variant="label" theme="neutral">
            Larger page size for dense data sets
          </Text>
        </div>

        <div className="p-4 border rounded-lg">
          <div className="flex justify-between items-center mb-3">
            <Text variant="subtext">5 items per page</Text>
            <Pagination offset={0} hasNext={true} limit={5} />
          </div>
          <Text variant="label" theme="neutral">
            Smaller page size for detailed content
          </Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Custom URL Parameter</h4>
      <div className="p-4 border rounded-lg">
        <div className="flex justify-between items-center mb-3">
          <Text variant="subtext">Using 'page' parameter instead of 'offset'</Text>
          <Pagination offset={10} hasNext={true} param="page" />
        </div>
        <Text variant="label" theme="neutral">
          This will update the 'page' URL parameter instead of 'offset'
        </Text>
      </div>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Pagination is commonly used with tables, lists, and search results.
        Here are typical implementation patterns for different scenarios.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Table Pagination</h4>
      <div className="border rounded-lg overflow-hidden">
        <div className="p-4 border-b bg-gray-50 dark:bg-gray-800">
          <Text weight="strong">User Management</Text>
        </div>
        <div className="p-4 space-y-2">
          <div className="flex justify-between py-2 border-b">
            <Text>John Doe</Text>
            <Text variant="subtext">john@example.com</Text>
          </div>
          <div className="flex justify-between py-2 border-b">
            <Text>Jane Smith</Text>
            <Text variant="subtext">jane@example.com</Text>
          </div>
          <div className="flex justify-between py-2">
            <Text>Bob Johnson</Text>
            <Text variant="subtext">bob@example.com</Text>
          </div>
        </div>
        <div className="p-4 border-t bg-gray-50 dark:bg-gray-800 flex justify-between items-center">
          <Text variant="subtext" theme="neutral">
            Showing 1-3 of 47 users
          </Text>
          <Pagination offset={0} hasNext={true} />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Search Results Pagination</h4>
      <div className="border rounded-lg p-4">
        <div className="flex justify-between items-center mb-4">
          <Text weight="strong">Search Results</Text>
          <Text variant="subtext" theme="neutral">
            142 results found
          </Text>
        </div>
        <div className="space-y-3 mb-4">
          <div className="p-3 border rounded">
            <Text weight="strong">Project Alpha</Text>
            <Text variant="subtext" className="mt-1">
              A comprehensive project management solution...
            </Text>
          </div>
          <div className="p-3 border rounded">
            <Text weight="strong">Beta Framework</Text>
            <Text variant="subtext" className="mt-1">
              Modern development framework for scalable...
            </Text>
          </div>
        </div>
        <div className="flex justify-center">
          <Pagination offset={20} hasNext={true} limit={10} />
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Implementation Guidelines:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Place pagination at the bottom of data sets for consistency</li>
        <li>Include result counts or range indicators when helpful</li>
        <li>Use appropriate page sizes based on content density</li>
        <li>Consider right-alignment for table footers</li>
        <li>Ensure pagination state persists across page navigation</li>
      </ul>
    </div>
  </div>
)
