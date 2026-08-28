import { ReleaseFiles, type TReleaseFileEntry } from './ReleaseFiles'

export default {
  title: 'Apps/Bundles/ReleaseFiles',
}

const entries: TReleaseFileEntry[] = [
  {
    path: 'components/api.toml',
    category: 'source',
    change: 'modified',
    current: { digest: 'sha256:current', size: 128 },
    previous: { digest: 'sha256:previous', size: 112 },
  },
  {
    path: 'policies/pass.rego',
    category: 'source',
    change: 'added',
    current: { digest: 'sha256:policy', size: 64 },
  },
  {
    path: 'package/components/api',
    category: 'artifact',
    change: 'modified',
    current: {
      digest: 'sha256:artifact-current',
      metadata: { kind: 'component', logical_name: 'api' },
      size: 2048,
    },
    previous: {
      digest: 'sha256:artifact-previous',
      metadata: { kind: 'component', logical_name: 'api' },
      size: 1900,
    },
  },
  {
    path: 'package/runtime/runner_binary/runner',
    category: 'runtime',
    change: 'unchanged',
    current: { digest: 'sha256:runner', size: 4096 },
    previous: { digest: 'sha256:runner', size: 4096 },
  },
]

export const Default = () => (
  <ReleaseFiles
    currentContent={{
      path: 'components/api.toml',
      content: 'name = "api"\nreplicas = 3\n',
      digest: 'sha256:current',
      media_type: 'application/toml',
      size: 128,
    }}
    entries={entries}
    onPackageChange={() => undefined}
    onSelect={() => undefined}
    packageOptions={[
      { id: 'package-1', platform: 'linux/amd64', status: 'active' },
    ]}
    previousContent={{
      path: 'components/api.toml',
      content: 'name = "api"\nreplicas = 2\n',
      digest: 'sha256:previous',
      media_type: 'application/toml',
      size: 112,
    }}
    selectedPackageId="package-1"
    selectedPath="components/api.toml"
  />
)
