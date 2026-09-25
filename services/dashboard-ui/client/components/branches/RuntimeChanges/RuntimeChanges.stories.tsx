import { RuntimeChanges } from './RuntimeChanges'

export default {
  title: 'Branches/RuntimeChanges',
}

const rows = [
  {
    buildId: 'bld53jcw9957zg1flwrimnumzk',
    componentName: 'img_event_proof',
    componentHref: '#',
    image:
      'us-west1-docker.pkg.dev/nuon-gcp-support/nuon-event-proof/clickhouse-server',
    resolvedTag: '1.0.1',
    digest:
      'sha256:45e09956dc667c5eff3583c9d94830261fb1ca0be10a0a7db36266edf5de9e1d',
    noOp: false,
    status: 'active',
    buildHref: '#',
  },
  {
    buildId: 'bldyvnpxiumrmfykj4ekct1tqm',
    componentName: 'img_no_change',
    componentHref: '#',
    image:
      'us-west1-docker.pkg.dev/nuon-gcp-support/nuon-event-proof/clickhouse-server',
    resolvedTag: '1.0.2',
    digest:
      'sha256:45e09956dc667c5eff3583c9d94830261fb1ca0be10a0a7db36266edf5de9e1d',
    noOp: true,
    status: 'active',
    buildHref: '#',
  },
  {
    buildId: 'bldfailedexample000000000',
    componentName: 'img_failed',
    componentHref: '#',
    image: 'ghcr.io/example/broken-image',
    resolvedTag: undefined,
    digest: undefined,
    noOp: false,
    status: 'error',
    buildHref: '#',
  },
  {
    buildId: 'blddockerbuildexample00000',
    componentName: 'api_docker_build',
    componentHref: '#',
    image: undefined,
    resolvedTag: undefined,
    digest: undefined,
    noOp: false,
    status: 'active',
    buildHref: '#',
  },
]

export const Default = () => <RuntimeChanges rows={rows} />
export const SingleNoOp = () => <RuntimeChanges rows={[rows[1]]} />
export const Empty = () => <RuntimeChanges rows={[]} />
