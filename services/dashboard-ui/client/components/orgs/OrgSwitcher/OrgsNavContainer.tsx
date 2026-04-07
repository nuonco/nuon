import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getOrgs } from '@/lib/ctl-api/orgs'
import { OrgsNav } from './OrgsNav'

const ENABLE_PAGINATION_COUNT = 6

export const OrgsNavContainer = () => {
  const [offset] = useState<number>(0)
  const [limit, setLimit] = useState<number>(10)
  const [searchTerm, setSearchTerm] = useState<string>('')

  const { data: orgs, isLoading } = useQuery({
    queryKey: ['orgs', { offset, limit, q: searchTerm }],
    queryFn: () => getOrgs({ offset, limit, q: searchTerm }),
  })

  return (
    <OrgsNav
      orgs={orgs}
      isLoading={isLoading}
      searchTerm={searchTerm}
      onSearchChange={setSearchTerm}
      onLoadMore={() => setLimit(limit + 10)}
      showSearch={orgs?.length > ENABLE_PAGINATION_COUNT || !!searchTerm}
      showLoadMore={orgs?.length > ENABLE_PAGINATION_COUNT && orgs?.length === limit}
    />
  )
}
