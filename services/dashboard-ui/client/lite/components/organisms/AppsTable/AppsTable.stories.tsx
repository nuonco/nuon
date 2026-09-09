import { DateTime } from 'luxon'
import type { TApp } from '@/types/ctl-api.types'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { AppsTable } from './AppsTable'

export default {
  title: 'lite/organisms/AppsTable',
}

const APPS: TApp[] = [
  {
    id: 'app_payments',
    name: 'Payments API',
    status_v2: { status: 'active' },
    runner_config: { cloud_platform: 'aws' },
    config_repo: 'example/payments',
    updated_at: DateTime.now().minus({ minutes: 8 }).toISO(),
  },
  {
    id: 'app_dashboard',
    name: 'Dashboard',
    status_v2: {
      status: 'applying',
      status_human_description: 'Applying the latest app configuration.',
    },
    runner_config: { cloud_platform: 'gcp' },
    config_repo: 'example/dashboard',
    updated_at: DateTime.now().minus({ hours: 2 }).toISO(),
  },
  {
    id: 'app_internal_tools',
    name: 'Internal tools',
    status_v2: { status: 'error' },
    runner_config: { cloud_platform: 'azure' },
    updated_at: DateTime.now().minus({ days: 1 }).toISO(),
  },
]

const props = {
  orgId: 'org_example',
  search: '',
  onSearchChange: () => {},
  offset: 0,
  pageSize: 20,
  hasNext: false,
  onOffsetChange: () => {},
}

export const Overview = () => (
  <ComponentDocs
    name="AppsTable"
    tier="organism"
    summary="The org's apps rendered through Table, with search and pagination bound to /v1/apps."
    use={[
      'Render it from the Apps page through AppsTableContainer.',
      'Read apps, search, and offset from the container so the URL stays the source of truth.',
    ]}
    avoid={[
      'Do not fetch apps inside the presentation component.',
      'Do not add filter dropdowns — /v1/apps only accepts a search term.',
    ]}
    rules={[
      'Search matches an app name or ID and resets the offset in the container.',
      'Pagination sits below Table, not in the toolbar.',
      'An error resolves to the empty state copy, never a thrown boundary.',
    ]}
    props={[
      {
        name: 'apps',
        type: 'TApp[]',
        description: 'Resolved apps for the current page.',
      },
      {
        name: 'orgId',
        type: 'string',
        description: 'Org segment used to build app hrefs.',
      },
      {
        name: 'search',
        type: 'string',
        description: 'Current search term from the URL.',
      },
      {
        name: 'onSearchChange',
        type: '(value: string) => void',
        description: 'Writes the search term and resets the offset.',
      },
      {
        name: 'offset',
        type: 'number',
        description: 'Row offset of the current page.',
      },
      {
        name: 'pageSize',
        type: 'number',
        description: 'Rows requested per page.',
      },
      {
        name: 'hasNext',
        type: 'boolean',
        description: 'Whether a full page came back, enabling next.',
      },
      {
        name: 'onOffsetChange',
        type: '(offset: number) => void',
        description: 'Moves the page window.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Renders Table skeletons on the cold load.',
      },
      {
        name: 'fetching',
        type: 'boolean',
        default: 'false',
        description: 'Disables pagination while a page is in flight.',
      },
      {
        name: 'error',
        type: 'unknown',
        description: 'Swaps the empty state to failure copy.',
      },
    ]}
  />
)

export const Default = () => <AppsTable {...props} apps={APPS} />

export const Loading = () => <AppsTable {...props} apps={[]} loading />

export const Empty = () => <AppsTable {...props} apps={[]} />

export const ErrorState = () => (
  <AppsTable {...props} apps={[]} error={new Error('Request failed')} />
)

export const NextPage = () => (
  <AppsTable {...props} apps={APPS} hasNext fetching />
)
