import { useState } from 'react'
import type { ReactNode } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Pagination } from './Pagination'

export default {
  title: 'lite/molecules/Pagination',
}

export const Overview = () => (
  <ComponentDocs
    name="Pagination"
    tier="molecule"
    summary="Controlled previous and next navigation for offset-paginated collections."
    use={[
      'Use below any collection backed by offset and has-next metadata.',
      'Connect onOffsetChange to useListQueryState.',
    ]}
    avoid={[
      'Do not read or write the URL from Pagination.',
      'Do not infer a total page count when the API only returns hasNext.',
    ]}
    rules={[
      'Previous is unavailable on the first page.',
      'Next is unavailable when hasNext is false.',
      'Both actions are unavailable while the next page is loading.',
    ]}
    props={[
      {
        name: 'offset',
        type: 'number',
        description: 'Zero-based API result offset.',
      },
      {
        name: 'pageSize',
        type: 'number',
        description: 'Fixed number of records requested by the collection.',
      },
      {
        name: 'hasNext',
        type: 'boolean',
        description: 'Whether the API reports another page.',
      },
      {
        name: 'onOffsetChange',
        type: '(offset: number) => void',
        description: 'Receives the previous or next page offset.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Disables navigation while a page is loading.',
      },
    ]}
  />
)

const Frame = ({ children }: { children: ReactNode }) => (
  <div className="flex min-h-32 items-center justify-center p-8">
    {children}
  </div>
)

export const FirstPage = () => (
  <Frame>
    <Pagination offset={0} pageSize={20} hasNext onOffsetChange={() => {}} />
  </Frame>
)

export const MiddlePage = () => (
  <Frame>
    <Pagination offset={40} pageSize={20} hasNext onOffsetChange={() => {}} />
  </Frame>
)

export const LastPage = () => (
  <Frame>
    <Pagination
      offset={80}
      pageSize={20}
      hasNext={false}
      onOffsetChange={() => {}}
    />
  </Frame>
)

export const Loading = () => (
  <Frame>
    <Pagination
      offset={20}
      pageSize={20}
      hasNext
      loading
      onOffsetChange={() => {}}
    />
  </Frame>
)

export const Interactive = () => {
  const [offset, setOffset] = useState(0)

  return (
    <Frame>
      <Pagination
        offset={offset}
        pageSize={20}
        hasNext={offset < 80}
        onOffsetChange={setOffset}
      />
    </Frame>
  )
}
