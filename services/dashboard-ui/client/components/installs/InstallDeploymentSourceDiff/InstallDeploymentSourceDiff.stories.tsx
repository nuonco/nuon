import { useState } from 'react'
import { PanelStory } from '@/components/__stories__/helpers'
import { Panel } from '@/components/surfaces/Panel'
import type { TInstallDeploymentSourceDiff } from '@/types'
import { InstallDeploymentSourceDiff } from './InstallDeploymentSourceDiff'
import { sourceDiff, sourceDiffContents } from './source-diff.fixtures'

export default { title: 'Features / Installs / Source changes' }

const SourceChangesContent = ({
  comparison = sourceDiff,
  isLoading = false,
  isError = false,
}: {
  comparison?: TInstallDeploymentSourceDiff
  isLoading?: boolean
  isError?: boolean
}) => {
  const [path, setPath] = useState<string>()
  return (
    <InstallDeploymentSourceDiff
      comparison={comparison}
      isLoading={isLoading}
      isError={isError}
      onRetry={() => {}}
      selectedPath={path}
      onSelectPath={setPath}
      fileComparison={
        path
          ? { ...comparison, selected_file: sourceDiffContents[path] }
          : undefined
      }
    />
  )
}

const SourceChangesStory = (
  props: Parameters<typeof SourceChangesContent>[0]
) => (
  <PanelStory label="View source changes">
    <Panel heading="Source changes" size="3/4" aria-label="Source changes">
      <SourceChangesContent {...props} />
    </Panel>
  </PanelStory>
)

export const Available = () => <SourceChangesStory />

export const Partial = () => (
  <SourceChangesStory
    comparison={{
      ...sourceDiff,
      partial: true,
      reasons: [
        'terraform_provider_closure_uninspected',
        'terraform_providers_not_locked',
        'GitHub inputs outside services/API_v2 may differ.',
      ],
    }}
  />
)

export const InitialDeployment = () => (
  <SourceChangesStory
    comparison={{
      ...sourceDiff,
      before: undefined,
      files: [{ path: 'new.yaml', change: 'added' }],
    }}
  />
)

export const CapturedFilesUnchanged = () => (
  <SourceChangesStory
    comparison={{ ...sourceDiff, files: [], partial: true }}
  />
)

export const Unavailable = () => (
  <SourceChangesStory
    comparison={{
      ...sourceDiff,
      status: 'unavailable',
      message:
        'The previously applied build has no captured source identity. Review its build metadata.',
      files: [],
    }}
  />
)

export const ImageDigests = () => (
  <SourceChangesStory
    comparison={{
      ...sourceDiff,
      status: 'unsupported',
      before: { id: 'bld-image-before', digest: `sha256:${'a'.repeat(64)}` },
      after: { id: 'bld-image-after', digest: `sha256:${'b'.repeat(64)}` },
      files: [],
    }}
  />
)

export const Loading = () => <SourceChangesStory isLoading />
export const Error = () => <SourceChangesStory isError />
