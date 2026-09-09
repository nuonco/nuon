import { useState } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Dropdown } from '../atoms/Dropdown'
import { Button } from '../atoms/Button'
import { Text } from '../atoms/Text'
import { SwitcherMenu } from './SwitcherMenu'

export default {
  title: 'lite/molecules/SwitcherMenu',
}

export const Overview = () => (
  <ComponentDocs
    name="SwitcherMenu"
    tier="molecule"
    summary="A searchable, paginated menu for switching the active resource."
    use={[
      'Switch between resources of one kind, such as organizations or app branches.',
      'Compose inside a Dropdown or as the content of a nested Menu submenu.',
      'Pass loadingContent so the loading rows match the shape of a real row.',
    ]}
    avoid={[
      'Do not use it for actions; every row is a navigation target.',
      'Do not mix resource kinds in one switcher.',
      'Do not add your own search debounce; the menu already debounces.',
    ]}
    rules={[
      'Search receives focus when the menu opens, and ArrowUp from the first row returns to it.',
      'Search commits on a debounce, so a keystroke is not a request.',
      'Load more appears only when another page exists.',
      'Loading, empty, and error are distinct states and never overlap.',
      'Rows fall back to a mono label when no content is supplied.',
    ]}
    props={[
      {
        name: 'items',
        type: 'ISwitcherMenuItem[]',
        description:
          'Rows to render. Each needs an id, label and href; content overrides the rendered body.',
      },
      {
        name: 'selectedId',
        type: 'string',
        description: 'Marks the active row as checked.',
      },
      {
        name: 'search',
        type: 'string',
        description: 'Committed search value owned by the container.',
      },
      {
        name: 'onSearchChange',
        type: '(value: string) => void',
        description: 'Receives the debounced search value.',
      },
      {
        name: 'onLoadMore',
        type: '() => void',
        description: 'Fetches the next page without closing the menu.',
      },
      {
        name: 'searchLabel',
        type: 'string',
        description: 'Accessible name for the search field.',
      },
      {
        name: 'searchPlaceholder',
        type: 'string',
        description: 'Placeholder text in the search field.',
      },
      {
        name: 'emptyTitle',
        type: 'string',
        description: 'Shown in place of the rows when nothing matches.',
      },
      {
        name: 'errorTitle',
        type: 'string',
        description: 'Shown in place of the rows when the fetch failed.',
      },
      {
        name: 'loadingContent',
        type: 'ReactNode',
        description:
          'Row body used for each loading row. Defaults to a single text bar.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Replaces the rows with five loading rows.',
      },
      {
        name: 'loadingMore',
        type: 'boolean',
        default: 'false',
        description: 'Puts the load more row into its pending state.',
      },
      {
        name: 'hasMore',
        type: 'boolean',
        default: 'false',
        description: 'Shows the load more row.',
      },
      {
        name: 'hasError',
        type: 'boolean',
        default: 'false',
        description: 'Replaces the rows with the error title.',
      },
    ]}
  />
)

const ITEMS = [
  { id: 'br_main', label: 'main', href: '/branches/br_main' },
  { id: 'br_release', label: 'release', href: '/branches/br_release' },
  { id: 'br_preview', label: 'preview', href: '/branches/br_preview' },
  { id: 'br_next', label: 'next', href: '/branches/br_next' },
  { id: 'br_legacy', label: 'legacy', href: '/branches/br_legacy' },
]

export const Default = () => {
  const [search, setSearch] = useState('')

  return (
    <div className="p-20">
      <Dropdown defaultOpen trigger={<Button>Switch branch</Button>}>
        <SwitcherMenu
          items={ITEMS.filter((item) => item.label.includes(search))}
          selectedId="br_main"
          search={search}
          onSearchChange={setSearch}
          onLoadMore={() => {}}
          searchLabel="Search branches"
          searchPlaceholder="Search branches..."
          emptyTitle="No branches found"
          errorTitle="Branches failed to load"
          hasMore
        />
      </Dropdown>
    </div>
  )
}

export const Loading = () => (
  <div className="p-20">
    <Dropdown defaultOpen trigger={<Button>Switch branch</Button>}>
      <SwitcherMenu
        items={[]}
        search=""
        onSearchChange={() => {}}
        onLoadMore={() => {}}
        searchLabel="Search branches"
        searchPlaceholder="Search branches..."
        emptyTitle="No branches found"
        errorTitle="Branches failed to load"
        loadingContent={<Text loading family="mono" loadingWidth={10} />}
        loading
      />
    </Dropdown>
  </div>
)

export const LoadingWithoutRowContent = () => (
  <div className="p-20">
    <Dropdown defaultOpen trigger={<Button>Switch branch</Button>}>
      <SwitcherMenu
        items={[]}
        search=""
        onSearchChange={() => {}}
        onLoadMore={() => {}}
        searchLabel="Search branches"
        searchPlaceholder="Search branches..."
        emptyTitle="No branches found"
        errorTitle="Branches failed to load"
        loading
      />
    </Dropdown>
  </div>
)

export const Empty = () => (
  <div className="p-20">
    <Dropdown defaultOpen trigger={<Button>Switch branch</Button>}>
      <SwitcherMenu
        items={[]}
        search="nothing"
        onSearchChange={() => {}}
        onLoadMore={() => {}}
        searchLabel="Search branches"
        searchPlaceholder="Search branches..."
        emptyTitle="No branches found"
        errorTitle="Branches failed to load"
      />
    </Dropdown>
  </div>
)

export const Error = () => (
  <div className="p-20">
    <Dropdown defaultOpen trigger={<Button>Switch branch</Button>}>
      <SwitcherMenu
        items={[]}
        search=""
        onSearchChange={() => {}}
        onLoadMore={() => {}}
        searchLabel="Search branches"
        searchPlaceholder="Search branches..."
        emptyTitle="No branches found"
        errorTitle="Branches failed to load"
        hasError
      />
    </Dropdown>
  </div>
)
