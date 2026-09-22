export default {
  title: 'Install Components/InstallImagesList',
}

import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { EmptyState } from '@/components/common/EmptyState'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { SearchInput } from '@/components/common/SearchInput'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import {
  InstallImagesList,
  type TInstallImageListItem,
} from './InstallImagesList'

const buildSummary = (
  <Card className="!p-4 !gap-4">
    <div className="grid gap-4 md:grid-cols-2">
      <LabeledValue label="Status">
        <Status status="active" />
      </LabeledValue>
      <LabeledValue label="Built">
        <Time time="2026-09-22T11:00:00Z" format="relative" variant="subtext" />
      </LabeledValue>
      <LabeledValue label="Source ref">
        <Text variant="subtext" family="mono">
          registry.example.com/acme/api:1.14.2
        </Text>
      </LabeledValue>
      <LabeledValue label="Resolved tag">
        <Text variant="subtext" family="mono">
          1.14.2
        </Text>
      </LabeledValue>
      <LabeledValue label="Digest" className="md:col-span-2">
        <ClickToCopy>
          <Text variant="subtext" family="mono" className="break-all">
            sha256:0000000000000000000000000000000000000000000000000000000000000000
          </Text>
        </ClickToCopy>
      </LabeledValue>
      <LabeledValue label="Rebuild">
        <Badge size="sm" variant="code" theme="neutral">
          no-op
        </Badge>
      </LabeledValue>
    </div>
    <Link href="#">View build</Link>
  </Card>
)

const buildAction = (
  <Button variant="secondary" size="sm">
    Build image
  </Button>
)

const image = (
  overrides: Partial<TInstallImageListItem>
): TInstallImageListItem => ({
  id: 'cmp-img-1',
  name: 'api',
  type: 'external_image',
  status: 'active',
  buildAction,
  image: buildSummary,
  ...overrides,
})

const images: TInstallImageListItem[] = [
  image({}),
  image({ id: 'cmp-img-2', name: 'worker', type: 'docker_build' }),
]

export const Default = () => (
  <InstallImagesList
    images={images}
    search={
      <SearchInput
        placeholder="Search by name or ID..."
        value=""
        onChange={() => {}}
      />
    }
    pagination={{ hasNext: true, offset: 0, limit: 10 }}
  />
)

export const NeverBuilt = () => (
  <InstallImagesList
    images={[
      image({
        image: (
          <EmptyState
            variant="table"
            size="sm"
            emptyTitle="No builds yet"
            emptyMessage="Build this component to publish an image."
          />
        ),
      }),
    ]}
  />
)

export const NoResults = () => (
  <InstallImagesList
    images={[]}
    filtered
    search={
      <SearchInput
        placeholder="Search by name or ID..."
        value="api"
        onChange={() => {}}
      />
    }
  />
)

export const Empty = () => <InstallImagesList images={[]} />

export const Loading = () => <InstallImagesList images={[]} loading />
