import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Avatar } from '@/components/common/Avatar'
import { Button } from '@/components/common/Button'
import { Link } from '@/components/common/Link'
import { SearchInput } from '@/components/common/SearchInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { getOrgs } from '@/lib/ctl-api/orgs'
import { OrgSummary } from './OrgSummary'

const LoadingOrgSummary = () => {
  return (
    <div className="flex gap-4 items-center p-2 w-full">
      <Avatar size="xl" isLoading />
      <div className="flex flex-col gap-1 transition-all w-full">
        <Skeleton height="14px" width="80%" />
        <Skeleton height="14px" width="40%" />
      </div>
    </div>
  )
}

export const OrgsNav = () => {
  const enablePaginationCount = 6
  const [offset] = useState<number>(0)
  const [limit, setLimit] = useState<number>(10)
  const [searchTerm, setSearchTerm] = useState<string>('')

  const { data: orgs, isLoading } = useQuery({
    queryKey: ['orgs', { offset, limit, q: searchTerm }],
    queryFn: () => getOrgs({ offset, limit, q: searchTerm }),
  })

  return (
    <div className="w-full">
      {orgs?.length > enablePaginationCount || searchTerm ? (
        <div className="p-2 w-full">
          <SearchInput
            labelClassName="md:!min-w-full md:!w-full"
            className="md:!min-w-full md:!w-full"
            placeholder="Search org by name..."
            value={searchTerm}
            onChange={(value) => setSearchTerm(value)}
          />
        </div>
      ) : null}
      {isLoading ? (
        Array.from({ length: 5 }).map((_, i) => <LoadingOrgSummary key={i} />)
      ) : orgs?.length ? (
        orgs?.map((o) => (
          <Link
            key={o?.id}
            className="!h-fit !block w-full"
            href={`/${o?.id}/apps`}
            variant="ghost"
          >
            <OrgSummary org={o} />
          </Link>
        ))
      ) : (
        <div className="flex flex-col items-center text-center w-full px-2 py-4">
          <Text variant="base" weight="strong">
            No org found
          </Text>
          <Text variant="subtext" theme="neutral">
            Clear your search and try again
          </Text>
        </div>
      )}
      {orgs?.length > enablePaginationCount && orgs?.length === limit ? (
        <Button
          className="w-full justify-center mt-4"
          onClick={() => setLimit(limit + 10)}
          variant="ghost"
        >
          Load more
        </Button>
      ) : null}
    </div>
  )
}
