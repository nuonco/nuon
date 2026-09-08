import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { Breadcrumb } from './Breadcrumb'

export default {
  title: 'lite/molecules/Breadcrumb',
}

const ITEMS = [
  { label: 'Acme', href: '/org' },
  { label: 'Installs', href: '/org/installs' },
  { label: 'Production', href: '/org/installs/production' },
  {
    label: 'Components',
    href: '/org/installs/production/components',
  },
  {
    label: 'API',
    href: '/org/installs/production/components/api',
  },
  { label: 'Deploy' },
]

export const Overview = () => (
  <ComponentDocs
    name="Breadcrumb"
    tier="molecule"
    summary="The current route hierarchy, rendered in the dashboard header."
    use={[
      'Render the items managed by useBreadcrumbs in DashboardHeader leading content.',
      'Keep the first and current page visible when space allows, with hidden ancestors in the overflow menu.',
    ]}
    avoid={[
      'Do not derive labels from URL segments.',
      'Do not place breadcrumbs inside page content.',
      'Do not give the current page an href.',
    ]}
    rules={[
      'Pages and nested route layouts own their breadcrumb content through useBreadcrumbs.',
      'An undefined label keeps the segment in place as loading text.',
      'The current page is plain text and ancestors are links.',
      'Overflow is available by keyboard and touch through a menu.',
    ]}
    props={[
      {
        name: 'items',
        type: 'IBreadcrumbItem[]',
        description: 'Ordered ancestors followed by the current page.',
      },
      {
        name: 'label',
        type: 'string',
        description: 'Segment text. Undefined renders its loading state.',
      },
      {
        name: 'href',
        type: 'string',
        description: 'Ancestor destination. Omit for the current page.',
      },
      {
        name: 'loadingWidth',
        type: 'number',
        description: 'Loading width in ch for an unresolved label.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="w-full p-8">
    <Breadcrumb
      items={[
        { label: 'Acme', href: '/org' },
        { label: 'Installs', href: '/org/installs' },
        { label: 'Production' },
      ]}
    />
  </div>
)

export const Loading = () => (
  <div className="w-full p-8">
    <Breadcrumb
      items={[
        { href: '/org', loadingWidth: 16 },
        { label: 'Installs', href: '/org/installs' },
        { loadingWidth: 12 },
      ]}
    />
  </div>
)

export const Overflow = () => (
  <div className="w-96 p-8">
    <Breadcrumb items={ITEMS} />
  </div>
)

export const Mobile = () => (
  <div className="w-64 p-4">
    <Breadcrumb items={ITEMS} />
  </div>
)
