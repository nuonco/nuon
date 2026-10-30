import { useMemo } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { CodeBlock } from '@/components/diffs/CodeBlock'
import { Diff } from '@/components/diffs/Diff'
import {
  DiffFileTree,
  type TDiffFileTreeEntry,
} from '@/components/diffs/DiffFileTree'
import type { TInstallDeploymentSourceDiff } from '@/types'
import { humanize } from '@/utils/string-utils'

type TSourceBuild = TInstallDeploymentSourceDiff['after']
type TSourceFile = NonNullable<TInstallDeploymentSourceDiff['files']>[number]

interface IInstallDeploymentSourceDiff {
  comparison?: TInstallDeploymentSourceDiff
  isLoading?: boolean
  isError?: boolean
  onRetry?: () => void
  selectedPath?: string
  onSelectPath: (path: string) => void
  fileComparison?: TInstallDeploymentSourceDiff
  isLoadingFile?: boolean
  isFileError?: boolean
  onRetryFile?: () => void
}

const SourceBuild = ({
  label,
  build,
  initial,
}: {
  label: string
  build?: TSourceBuild
  initial?: boolean
}) => (
  <div className="flex min-w-0 flex-col gap-3">
    <Text weight="strong">{label}</Text>
    {build ? (
      <>
        {[
          ['Build ID', build.id],
          ['Repository', build.repository],
          ['Root', build.root],
          ['Pinned commit SHA', build.commit],
          ['Image digest', build.digest],
        ].map(([field, value]) =>
          value ? (
            <LabeledValue key={field} label={field}>
              <Text family="mono" variant="subtext" className="break-all">
                {value}
              </Text>
            </LabeledValue>
          ) : null
        )}
        {build.coverage ? (
          <LabeledValue label="Coverage">
            <Text variant="subtext">{humanize(build.coverage)}</Text>
          </LabeledValue>
        ) : null}
      </>
    ) : (
      <Text variant="subtext" theme="neutral">
        {initial
          ? 'Initial deployment — no previously applied build.'
          : 'Build metadata unavailable.'}
      </Text>
    )}
  </div>
)

const LoadingFailed = ({
  file = false,
  onRetry,
}: {
  file?: boolean
  onRetry?: () => void
}) => (
  <Banner theme="error">
    <div className="flex flex-col gap-3">
      <Text weight="strong">
        {file ? 'File loading failed' : 'Source changes loading failed'}
      </Text>
      <Text>Try again to load {file ? 'this file' : 'this comparison'}.</Text>
      {onRetry ? <Button onClick={onRetry}>Try again</Button> : null}
    </div>
  </Banner>
)

const sourceLanguage = (path: string) => {
  const name = path.toLowerCase()
  if (name.endsWith('dockerfile')) return 'docker'
  return name.split('.').at(-1)
}

const SelectedFile = ({
  file,
  comparison,
  isLoading,
  isError,
  onRetry,
}: {
  file: TSourceFile
  comparison?: TInstallDeploymentSourceDiff
  isLoading?: boolean
  isError?: boolean
  onRetry?: () => void
}) => {
  const selected = comparison?.selected_file
  if (isError) return <LoadingFailed file onRetry={onRetry} />
  if (isLoading)
    return (
      <div role="status" aria-label="Loading file">
        <Loading />
      </div>
    )
  if (selected?.path === file.path && selected.unavailable_reason)
    return <Banner theme="warn">{selected.unavailable_reason}</Banner>
  if (
    comparison?.status !== 'available' ||
    selected?.path !== file.path ||
    (file.change !== 'added' && typeof selected.before !== 'string') ||
    (file.change !== 'removed' && typeof selected.after !== 'string')
  )
    return (
      <Banner theme="warn">
        {comparison?.message ||
          'File contents unavailable. Select another file or try again later.'}
      </Banner>
    )

  const before = selected.before ?? ''
  const after = selected.after ?? ''
  const modeChanged =
    file.before_executable !== undefined &&
    file.after_executable !== undefined &&
    file.before_executable !== file.after_executable

  return (
    <div className="flex flex-col gap-3">
      {modeChanged ? (
        <Banner theme="info">
          Executable bit changed:{' '}
          {file.before_executable ? 'executable' : 'not executable'} →{' '}
          {file.after_executable ? 'executable' : 'not executable'}.
        </Banner>
      ) : null}
      {before === after ? (
        <>
          <Text variant="subtext" theme="neutral">
            {file.change === 'added'
              ? 'Empty file added.'
              : file.change === 'removed'
                ? 'Empty file removed.'
                : 'File contents unchanged.'}
          </Text>
          <CodeBlock
            value={after}
            filename={file.path}
            language={sourceLanguage(file.path)}
          />
        </>
      ) : (
        <Diff
          before={before}
          after={after}
          filename={file.path}
          language={sourceLanguage(file.path)}
          view="unified"
        />
      )}
    </div>
  )
}

export const InstallDeploymentSourceDiff = ({
  comparison,
  isLoading,
  isError,
  onRetry,
  selectedPath,
  onSelectPath,
  fileComparison,
  isLoadingFile,
  isFileError,
  onRetryFile,
}: IInstallDeploymentSourceDiff) => {
  const files = comparison?.files ?? []
  const entries = useMemo<TDiffFileTreeEntry[]>(
    () =>
      files.map(({ path, change }) => ({
        path,
        status: change === 'removed' ? 'deleted' : change,
      })),
    [files]
  )
  const selected = files.find(({ path }) => path === selectedPath)

  if (isError) return <LoadingFailed onRetry={onRetry} />
  if (isLoading)
    return (
      <div role="status" aria-label="Loading source changes">
        <Loading variant="large" />
      </div>
    )
  if (!comparison) return <LoadingFailed onRetry={onRetry} />

  return (
    <div className="@container flex flex-col gap-4">
      <Text family="mono" weight="strong">
        {comparison.name}
      </Text>
      <Text variant="subtext" theme="neutral">
        Previously applied build → selected build. This comparison uses the
        frozen build sources for this deployment.
      </Text>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <SourceBuild
          label="Previously applied build"
          build={comparison.before}
          initial={comparison.status === 'available' && !comparison.before}
        />
        <SourceBuild label="Selected build" build={comparison.after} />
      </div>
      {comparison.partial || comparison.reasons?.length ? (
        <Banner theme="warn" role="note" aria-label="Source coverage">
          <div className="flex flex-col gap-2">
            {comparison.partial ? (
              <Text variant="subtext">
                Only captured source files are compared. Other build inputs may
                be uncaptured; this comparison cannot prove all source is
                unchanged.
              </Text>
            ) : null}
            {comparison.reasons?.length ? (
              <ul className="list-disc pl-4">
                {comparison.reasons.map((reason) => (
                  <li key={reason}>
                    <Text variant="subtext">
                      {/^[a-z][a-z0-9]*(?:[_-][a-z0-9]+)+$/.test(reason)
                        ? humanize(reason)
                        : reason}
                    </Text>
                  </li>
                ))}
              </ul>
            ) : null}
          </div>
        </Banner>
      ) : null}
      {comparison.status !== 'available' ? (
        <Banner theme="warn">
          <div className="flex flex-col gap-2">
            <Text weight="strong">
              {comparison.status === 'unsupported'
                ? 'Text comparison unsupported'
                : 'Source comparison unavailable'}
            </Text>
            <Text>
              {comparison.message ||
                (comparison.status === 'unsupported'
                  ? 'This build uses an image. Compare the image digests above; a text diff is not available.'
                  : 'Source could not be compared. Review the build metadata or try again later.')}
            </Text>
          </div>
        </Banner>
      ) : (
        <>
          {comparison.message ? (
            <Text theme="neutral">{comparison.message}</Text>
          ) : null}
          {!files.length ? (
            <Text theme="neutral">
              Captured source files unchanged. Review the build metadata for
              other input changes.
            </Text>
          ) : (
            <div className="flex flex-col gap-3">
              <Text weight="strong">Changed files</Text>
              <div className="flex flex-col overflow-hidden rounded-md border @3xl:flex-row">
                <div className="h-72 overflow-auto border-b bg-code pt-4 @3xl:h-144 @3xl:w-1/3 @3xl:border-b-0 @3xl:border-r">
                  <DiffFileTree
                    files={entries}
                    selectedPath={selectedPath}
                    onSelect={onSelectPath}
                  />
                </div>
                <div className="flex min-w-0 flex-col gap-3 p-4 @3xl:w-2/3">
                  {selected ? (
                    <>
                      <Text family="mono" className="break-all">
                        {selected.path}
                      </Text>
                      <SelectedFile
                        key={selected.path}
                        file={selected}
                        comparison={fileComparison}
                        isLoading={isLoadingFile}
                        isError={isFileError}
                        onRetry={onRetryFile}
                      />
                    </>
                  ) : (
                    <Text theme="neutral">
                      Select a changed file to view its diff.
                    </Text>
                  )}
                </div>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
