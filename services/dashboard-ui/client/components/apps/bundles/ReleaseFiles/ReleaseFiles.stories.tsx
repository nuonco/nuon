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
    onSelect={() => undefined}
    previousContent={{
      path: 'components/api.toml',
      content: 'name = "api"\nreplicas = 2\n',
      digest: 'sha256:previous',
      media_type: 'application/toml',
      size: 112,
    }}
    selectedPath="components/api.toml"
  />
)

export const FileTooLargeToPreview = () => (
  <ReleaseFiles
    entries={[
      {
        path: 'policies/large.rego',
        category: 'source',
        change: 'added',
        current: { digest: 'sha256:large', size: 5 * 1024 * 1024 + 1 },
      },
    ]}
    onSelect={() => undefined}
    selectedPath="policies/large.rego"
  />
)
