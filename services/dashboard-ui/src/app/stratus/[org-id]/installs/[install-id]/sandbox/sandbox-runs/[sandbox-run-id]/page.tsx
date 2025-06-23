import type { FC } from 'react'
import {
  BackLink,
  Header,
  HeadingGroup,
  ScrollableDiv,
  Section,
  Text,
} from '@/stratus/components'
import type { IPageProps, TSandboxRun } from '@/types'
import { nueQueryData } from '@/utils'

const SandboxRunPage: FC<
  IPageProps<'org-id' | 'install-id' | 'sandbox-run-id'>
> = async ({ params }) => {
  const {
    ['install-id']: installId,
    ['org-id']: orgId,
    ['sandbox-run-id']: sandboxRunId,
  } = await params
  const { data, error } = await nueQueryData<TSandboxRun>({
    orgId,
    path: `installs/sandbox-runs/${sandboxRunId}`,
  })

  return (
    <ScrollableDiv>
      <Header>
        <HeadingGroup>
          <BackLink />
          <Text variant="h3" weight="strong">
            {data?.run_type}
          </Text>
          <Text family="mono" variant="subtext" theme="muted">
            {data?.id}
          </Text>
        </HeadingGroup>
      </Header>
      <Section>Content</Section>
    </ScrollableDiv>
  )
}

export default SandboxRunPage
