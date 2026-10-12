import { useEffect, useMemo, useRef, type CSSProperties } from 'react'
import type { GitStatusEntry } from '@pierre/trees'
import { FileTree, useFileTree } from '@pierre/trees/react'

export type TDiffFileTreeStatus = 'added' | 'modified' | 'deleted' | 'unchanged'

export type TDiffFileTreeEntry = {
  path: string
  status: TDiffFileTreeStatus
}

export interface IDiffFileTree {
  files: TDiffFileTreeEntry[]
  selectedPath?: string
  onSelect: (path: string) => void
  // A new nonce re-reveals the same path, e.g. when a summary row is clicked twice.
  focus?: { path: string; nonce: number }
  search?: boolean
}

const TREE_STYLE = {
  height: '100%',
  '--trees-bg-override': 'var(--code-bg)',
  '--trees-border-color-override': 'var(--border-color)',
  '--trees-fg-override': 'var(--foreground)',
  '--trees-font-family-override': 'var(--font-mono)',
  '--trees-font-size-override': '12px',
  '--trees-padding-inline-override': '16px',
  '--trees-selected-bg-override': 'var(--background-neutral)',
  '--trees-git-added-color-override': 'var(--diff-add-gutter)',
  '--trees-git-deleted-color-override': 'var(--diff-remove-gutter)',
  '--trees-status-added-override': 'var(--diff-add-gutter)',
  '--trees-status-deleted-override': 'var(--diff-remove-gutter)',
} as CSSProperties

export const DiffFileTree = ({
  files,
  selectedPath,
  onSelect,
  focus,
  search = true,
}: IDiffFileTree) => {
  const filesRef = useRef(files)
  filesRef.current = files
  const onSelectRef = useRef(onSelect)
  onSelectRef.current = onSelect
  const treeSelectedPath = useRef<string | undefined>(undefined)
  const resetKeyRef = useRef<string | undefined>(undefined)

  const paths = useMemo(() => files.map(({ path }) => path), [files])
  const gitStatus = useMemo<GitStatusEntry[]>(
    () =>
      files.flatMap(({ path, status }) =>
        status === 'unchanged' ? [] : [{ path, status }]
      ),
    [files]
  )

  const { model } = useFileTree({
    density: 'compact',
    fileTreeSearchMode: 'expand-matches',
    flattenEmptyDirectories: true,
    gitStatus,
    icons: 'standard',
    initialExpansion: 'open',
    initialSelectedPaths: selectedPath ? [selectedPath] : [],
    onSelectionChange: (selectedPaths) => {
      const path = selectedPaths.at(-1)
      treeSelectedPath.current = path
      if (path && filesRef.current.some((file) => file.path === path)) {
        onSelectRef.current(path)
      }
    },
    paths,
    search,
  })

  // `files` changes identity whenever lazy content fetches resolve, but the
  // path set usually stays the same. resetPaths rebuilds the store with
  // initialExpansion applied, so only reset when paths or statuses change.
  const pathsKey = paths.join('\n')
  const gitStatusKey = gitStatus
    .map(({ path, status }) => `${path}:${status}`)
    .join('\n')
  const pathsRef = useRef(paths)
  pathsRef.current = paths
  const gitStatusRef = useRef(gitStatus)
  gitStatusRef.current = gitStatus

  useEffect(() => {
    if (resetKeyRef.current === `${pathsKey}\n${gitStatusKey}`) return
    resetKeyRef.current = `${pathsKey}\n${gitStatusKey}`
    model.resetPaths(pathsRef.current)
    model.setGitStatus(gitStatusRef.current)
  }, [gitStatusKey, model, pathsKey])

  // Sync a parent-driven selection into the tree. Tree-driven selections
  // already live in the model; scrolling for them would fight the user.
  useEffect(() => {
    if (!selectedPath || selectedPath === treeSelectedPath.current) return
    model.getItem(selectedPath)?.select()
  }, [model, selectedPath])

  useEffect(() => {
    if (!focus) return
    if (!filesRef.current.some((file) => file.path === focus.path)) return
    // Expand ancestors so scrollToPath can see the target even when its
    // folder is collapsed (canonical directory paths end with '/').
    const segments = focus.path.split('/')
    for (let i = 1; i < segments.length; i++) {
      const dir = model.getItem(`${segments.slice(0, i).join('/')}/`)
      if (dir != null && 'expand' in dir) dir.expand()
    }
    model.getItem(focus.path)?.select()
    model.scrollToPath(focus.path, { offset: 'nearest' })
  }, [focus, model])

  // The tree overwrites its host's class attribute, so size it via style.
  return <FileTree model={model} style={TREE_STYLE} />
}
