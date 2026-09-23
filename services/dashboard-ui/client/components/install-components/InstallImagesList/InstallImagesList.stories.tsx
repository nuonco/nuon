export default {
  title: 'Install Components/InstallImagesList',
}

import { Button } from '@/components/common/Button'
import { SearchInput } from '@/components/common/SearchInput'
import type { TComponentBuild, TDeploy } from '@/types'
import { InstallImageSummary } from './InstallImageSummary'
import {
  InstallImagesList,
  type TInstallImageListItem,
} from './InstallImagesList'

const sync = {
  id: 'dpl-img-1',
  status_v2: { status: 'active' },
  created_at: '2026-09-22T14:00:00Z',
  updated_at: '2026-09-22T14:04:00Z',
  install_deploy_type: 'apply',
} as TDeploy

const build = {
  id: 'bld-img-1',
  status_v2: { status: 'active' },
  created_at: '2026-09-22T11:00:00Z',
  resolved_at: '2026-09-22T11:02:00Z',
  source_ref: 'registry.example.com/acme/api:1.14.2',
  resolved_tag: '1.14.2',
  source_digest:
    'sha256:0000000000000000000000000000000000000000000000000000000000000000',
  no_op: true,
  vcs_connection_commit: {
    sha: '7a3c91e4b2d8f056c19a4e7b3d2058f46a1c9b2e',
    message: 'Pin acme-api image to 1.14.2',
    author_name: 'Ada Lovelace',
  },
} as TComponentBuild

const summary = (
  <InstallImageSummary
    deploy={sync}
    syncHref="#"
    build={build}
    buildHref="#"
  />
)

const actions = (
  <Button variant="secondary" size="sm">
    More
  </Button>
)

const image = (
  overrides: Partial<TInstallImageListItem>
): TInstallImageListItem => ({
  id: 'cmp-img-1',
  name: 'api',
  type: 'external_image',
  status: 'active',
  actions,
  image: summary,
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
        image: <InstallImageSummary />,
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
