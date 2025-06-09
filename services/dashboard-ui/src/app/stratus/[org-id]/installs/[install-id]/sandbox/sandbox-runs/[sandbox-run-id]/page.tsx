import type { FC } from 'react'
import {
  Header,
  HeaderGroup,
  ScrollableContent,
  Section,
  Text,
} from '@/stratus/components'
import type { IPageProps, TSandboxRun } from '@/types'
import { nueQueryData } from '@/utils'

const SandboxRunPage: FC<
  IPageProps<'org-id' | 'install-id' | 'sandbox-run-id'>
> = async ({ params }) => {
  const orgId = params?.['org-id']
  const installId = params?.['install-id']
  const sandboxRunId = params?.['sandbox-run-id']
  const { data, error } = await nueQueryData<TSandboxRun>({
    orgId,
    path: `installs/sandbox-runs/${sandboxRunId}`,
  })

  return (
    <ScrollableContent>
      <Header>
        <HeaderGroup>
          <Text variant="h3" weight="strong">
            {data?.run_type}
          </Text>
          <Text family="mono" variant="subtext" theme="muted">
            {data?.id}
          </Text>
        </HeaderGroup>
      </Header>
      <Section>Content</Section>
    </ScrollableContent>
  )
}

export default SandboxRunPage
