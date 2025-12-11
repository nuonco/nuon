'use client'

import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ID } from '@/components/common/ID'
import { ContextTooltip } from '@/components/common/ContextTooltip'
import type { TBuild } from '@/types'

interface CommitDetailsProps {
  commit: TBuild['vcs_connection_commit']
}

export const CommitDetails = ({ commit }: CommitDetailsProps) => {
  if (!commit) {
    return null
  }

  return (
    <ContextTooltip
      title="Commit details"
      items={[
        {
          id: commit?.author_email,
          title: 'Author',
          subtitle: commit?.author_name,
        },
        {
          id: commit?.id,
          title: 'Message',
          subtitle: (
            <Text
              variant="subtext"
              theme="neutral"
              className="!block !max-w-[180px] !truncate"
            >
              {commit?.message}
            </Text>
          ),
        },
        {
          id: commit?.vcs_connection_id,
          title: 'Date',
          subtitle: (
            <Time
              variant="subtext"
              theme="neutral"
              time={commit?.created_at}
            />
          ),
        },
      ]}
    >
      <ID>
        <span className="!block !max-w-[60px] !truncate">
          {commit?.sha}
        </span>
      </ID>
    </ContextTooltip>
  )
}