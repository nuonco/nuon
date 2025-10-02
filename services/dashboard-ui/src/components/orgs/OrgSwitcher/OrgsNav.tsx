'use client'

import classNames from 'classnames'
import React, { type FC, useState } from 'react'
import { TestTubeIcon } from '@phosphor-icons/react'
import NextLink from 'next/link'
import { Button } from '@/components/Button'
import { StatusBadge } from '@/components/Status'
import { Skeleton } from '@/components/Skeleton'
import { Text } from '@/components/Typography'
import type { TOrg } from '@/types'
import { SearchInput } from '@/components/SearchInput'
import { useQuery } from '@/hooks/use-query'
import { buildQueryParams } from '@/utils/build-query-params'
import { OrgAvatar } from './OrgAvatar'

export interface IOrgsNav {
  orgs: Array<TOrg>
}

export const OrgsNav: FC<IOrgsNav> = ({ orgs: initOrgs }) => {
  const enablePaginationCount = 6
  const [offset, setOffset] = useState<number>(0)
  const [limit, setLimit] = useState<number>(10)
  const [searchTerm, setSearchTerm] = useState<string>('')
  const queryParams = buildQueryParams({ offset, limit, q: searchTerm })

  const {
    data: orgs,
    isLoading,
    error,
    headers,
  } = useQuery<TOrg[]>({
    path: `/api/orgs${queryParams}`,
    initData: initOrgs,
  })

  return (
    <div className="flex flex-col gap-4">
      <Text className="px-4" variant="med-14">
        Organizations
      </Text>

      {orgs?.length > enablePaginationCount || searchTerm ? (
        <div className="px-4 w-full">
          <SearchInput
            labelClassName="md:!min-w-full md:!w-full"
            className="md:!min-w-full md:!w-full"
            placeholder="Search org name..."
            value={searchTerm}
            onChange={setSearchTerm}
          />
        </div>
      ) : null}

      <nav className="flex flex-col gap-0 px-1">
        {isLoading ? (
          <Skeleton
            className="flex flex-col items-center gap-2"
            lines={limit || 3}
            height="52px"
            width="230px"
          />
        ) : orgs?.length ? (
          orgs.map((org) => (
            <NextLink
              className="flex items-center justify-start gap-4 rounded-md p-2 hover:bg-cool-grey-600/20"
              key={org.id}
              href={`/${org.id}/apps`}
            >
              <OrgAvatar name={org.name} logoURL={org.logo_url} />
              <span>
                <Text
                  className="break-all text-md font-medium leading-normal mb-1 !flex-nowrap"
                  title={
                    org.sandbox_mode ? 'Org is in sandbox mode' : undefined
                  }
                >
                  {org.sandbox_mode && <TestTubeIcon className="text-sm" />}
                  <span
                    className={classNames('', {
                      'truncate !inline max-w-[140px]': org.name.length >= 16,
                    })}
                  >
                    {org.name}
                  </span>
                </Text>
                <StatusBadge status={org.status} isWithoutBorder />
              </span>
            </NextLink>
          ))
        ) : (
          <div className="flex flex-col gap-2 items-center text-center w-full px-2 py-4">
            <Text variant="med-14">No org found</Text>
            <Text variant="reg-12" isMuted>
              Clear your search and try again
            </Text>
          </div>
        )}
        {orgs.length > enablePaginationCount &&
        headers?.['x-nuon-page-next'] === 'true' ? (
          <Button
            onClick={() => {
              setLimit(limit + 10)
            }}
            className="w-full text-[12px] my-2 !py-3"
            variant="ghost"
          >
            Load more
          </Button>
        ) : null}
      </nav>
    </div>
  )
}
