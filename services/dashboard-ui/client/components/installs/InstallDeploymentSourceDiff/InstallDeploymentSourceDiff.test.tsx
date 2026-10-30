import type { ReactNode } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, beforeEach, expect, test } from 'bun:test'
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { BuildChangeDetail } from '@/components/branches/ConfigChanges/BuildChangeDetail'
import { ConfigChangesViewer } from '@/components/branches/ConfigChanges/ConfigChangesViewer'
import type { TTemplateBuildChange } from '@/components/branches/ConfigChanges/config-changes'
import {
  DashboardPreferencesContext,
  type IDashboardPreferencesContext,
} from '@/providers/dashboard-preferences-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { InstallDeploymentSourceDiff } from './InstallDeploymentSourceDiff'
import { InstallDeploymentSourceDiffPanel } from './InstallDeploymentSourceDiffContainer'
import { sourceDiff, sourceDiffContents } from './source-diff.fixtures'

const originalFetch = globalThis.fetch
const clients: QueryClient[] = []
const shadowActiveElement = Object.getOwnPropertyDescriptor(
  ShadowRoot.prototype,
  'activeElement'
)!
beforeEach(() => {
  // Happy DOM 18 throws instead of returning null when a different shadow root owns focus.
  Object.defineProperty(ShadowRoot.prototype, 'activeElement', {
    ...shadowActiveElement,
    get() {
      if (document.activeElement !== this.host) return null
      return shadowActiveElement.get!.call(this)
    },
  })
})
afterEach(() => {
  cleanup()
  clients.splice(0).forEach((client) => client.clear())
  globalThis.fetch = originalFetch
  Object.defineProperty(
    ShadowRoot.prototype,
    'activeElement',
    shadowActiveElement
  )
})

const target = {
  orgId: 'org-acme',
  installId: 'ins-acme',
  workflowId: 'wf-deploy',
  kind: 'component' as const,
  name: 'acme-api',
}
const build: TTemplateBuildChange = {
  id: 'bld-selected',
  name: 'acme-api',
  kind: 'component',
  status: 'active',
  changeReason: 'source_changed',
  href: '/org-acme/apps/app-acme/components/cmp-api/builds/bld-selected',
}

const renderUI = (node: ReactNode) => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  clients.push(client)
  const Wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>
      <DashboardPreferencesContext.Provider
        value={{ diffWrap: 'scroll' } as IDashboardPreferencesContext}
      >
        <MemoryRouter>
          <SurfacesProvider>{children}</SurfacesProvider>
        </MemoryRouter>
      </DashboardPreferencesContext.Provider>
    </QueryClientProvider>
  )
  return render(node, { wrapper: Wrapper })
}

const selectFile = async (name: string) => {
  let row: HTMLElement
  await waitFor(() => {
    const root = document.querySelector('file-tree-container')?.shadowRoot
    expect(root).toBeTruthy()
    row = within(root as unknown as HTMLElement).getByRole('treeitem', {
      name: new RegExp(name.replace('.', '\\.')),
    })
  })
  await act(async () => {
    fireEvent.click(row!)
  })
}

const codeText = () =>
  Array.from(document.querySelectorAll('diffs-container'))
    .map((element) => element.shadowRoot?.textContent ?? '')
    .join('\n')

test('branch build rows keep View build but do not offer an install source comparison', () => {
  renderUI(<BuildChangeDetail build={build} />)
  expect(screen.getByRole('link', { name: 'View build' })).toHaveAttribute(
    'href',
    build.href
  )
  expect(screen.getByText('acme-api')).toBeTruthy()
  expect(
    screen.queryByRole('button', { name: /View source changes/ })
  ).toBeNull()
})

test('opens from a build row, fetches files on demand, and clears old contents while switching or failing', async () => {
  const calls: URL[] = []
  let finishBinary: (response: Response) => void
  let failRemoved = true
  globalThis.fetch = (async (input: string, init: RequestInit) => {
    const url = new URL(input, 'http://example.test')
    calls.push(url)
    expect(init.headers['X-Nuon-Org-ID']).toBe(target.orgId)
    const path = url.searchParams.get('path')
    if (path === 'logo.png')
      return new Promise<Response>((resolve) => {
        finishBinary = resolve
      })
    if (path === 'old.yaml' && failRemoved)
      return Response.json({ error: 'Access failed' }, { status: 503 })
    return Response.json({
      ...sourceDiff,
      selected_file: path ? sourceDiffContents[path] : undefined,
    })
  }) as typeof fetch

  const { container } = renderUI(
    <ConfigChangesViewer
      sections={[]}
      builds={[{ ...build, sourceDiffTarget: target }]}
    />
  )
  expect(calls).toHaveLength(0)
  const identity = screen.getByRole('button', {
    name: 'View source changes for acme-api',
  })
  expect(identity.textContent).toBe('acme-api')
  const row = container.querySelector(
    '[data-config-change="components/acme-api"]'
  )!
  expect(row.querySelector('[aria-expanded]')).toBeNull()
  expect(within(row as HTMLElement).getByText('Source changed')).toBeTruthy()
  expect(
    within(row as HTMLElement).getByRole('link', { name: 'View build' })
  ).toBeTruthy()
  fireEvent.click(identity)
  expect(
    await screen.findByRole('complementary', { name: 'Source changes' })
  ).toBeTruthy()
  expect(await screen.findByText('bld-previous')).toBeTruthy()
  expect(screen.getByText('bld-selected')).toBeTruthy()
  expect(screen.getAllByText('acme/example')).toHaveLength(2)
  expect(screen.getAllByText('services/api')).toHaveLength(2)
  expect(screen.getByText(sourceDiff.before!.commit!)).toBeTruthy()
  expect(screen.getByText(sourceDiff.after!.commit!)).toBeTruthy()
  expect(calls).toHaveLength(1)
  expect(calls[0].pathname).toBe(
    '/v1/installs/ins-acme/deployments/wf-deploy/source-diff'
  )
  expect(calls[0].searchParams.get('kind')).toBe('component')
  expect(calls[0].searchParams.get('name')).toBe('acme-api')
  expect(calls[0].searchParams.has('path')).toBe(false)

  await selectFile('settings.yaml')
  await waitFor(() => expect(codeText()).toContain('replicas'))
  expect(calls.at(-1)?.searchParams.get('path')).toBe('settings.yaml')
  await selectFile('logo.png')
  expect(
    await screen.findByRole('status', { name: 'Loading file' })
  ).toBeTruthy()
  expect(screen.getByText('logo.png')).toBeTruthy()
  expect(codeText()).not.toContain('replicas')
  await act(async () => {
    finishBinary!(
      Response.json({
        ...sourceDiff,
        selected_file: sourceDiffContents['logo.png'],
      })
    )
  })
  expect(
    await screen.findByText(sourceDiffContents['logo.png'].unavailable_reason!)
  ).toBeTruthy()
  expect(codeText()).not.toContain('replicas')
  expect(screen.queryByText('File contents unchanged.')).toBeNull()

  await selectFile('old.yaml')
  expect(await screen.findByText('File loading failed')).toBeTruthy()
  expect(
    screen.queryByText(sourceDiffContents['logo.png'].unavailable_reason!)
  ).toBeNull()
  expect(codeText()).toBe('')
  failRemoved = false
  fireEvent.click(screen.getByRole('button', { name: 'Try again' }))
  await waitFor(() => expect(codeText()).toContain('legacy'))
  await selectFile('new.yaml')
  await waitFor(() => expect(codeText()).toContain('enabled'))
  expect(codeText()).not.toContain('legacy')
})

test('keeps config changes visible beside the source identity trigger without a second disclosure', async () => {
  const { container } = renderUI(
    <ConfigChangesViewer
      sections={[
        {
          name: 'Components',
          sectionKey: 'components',
          additions: 0,
          removals: 0,
          changed: 1,
          grouped: true,
          fields: [],
          entities: [
            {
              name: build.name,
              op: 'change',
              fields: [{ key: 'replicas', op: 'change', diff: '1 -> 4' }],
            },
          ],
        },
      ]}
      builds={[
        {
          ...build,
          changeReason: 'source_and_config',
          sourceDiffTarget: target,
        },
      ]}
    />
  )
  expect(
    screen.getByRole('button', { name: 'View source changes for acme-api' })
  ).toBeTruthy()
  const row = container.querySelector(
    '[data-config-change="components/acme-api"]'
  )!
  expect(row.querySelector('[aria-expanded]')).toBeNull()
  await waitFor(() => {
    expect(codeText()).toContain('replicas = 1')
    expect(codeText()).toContain('replicas = 4')
  })
})

test('distinguishes initial, partial, unavailable, and image-only comparisons without inventing text diffs', () => {
  const props = { onSelectPath: () => {} }
  const { rerender } = renderUI(
    <InstallDeploymentSourceDiff
      {...props}
      comparison={{
        ...sourceDiff,
        before: undefined,
        files: [],
        partial: true,
        reasons: [
          'terraform_provider_closure_uninspected',
          'terraform_providers_not_locked',
          'GitHub inputs outside services/API_v2 may differ.',
        ],
      }}
    />
  )
  expect(screen.getByText(/Initial deployment/)).toBeTruthy()
  expect(screen.getByText(/Captured source files unchanged/)).toBeTruthy()
  const coverage = within(screen.getByRole('note', { name: 'Source coverage' }))
  expect(
    coverage.getByText(/cannot prove all source is unchanged/)
  ).toBeTruthy()
  expect(
    coverage.getByText('Terraform provider closure uninspected')
  ).toBeTruthy()
  expect(coverage.getByText('Terraform providers not locked')).toBeTruthy()
  expect(
    coverage.getByText('GitHub inputs outside services/API_v2 may differ.')
  ).toBeTruthy()

  rerender(
    <InstallDeploymentSourceDiff
      {...props}
      comparison={{
        ...sourceDiff,
        before: undefined,
        status: 'unavailable',
        message: 'Source identity was not captured.',
        files: [],
      }}
    />
  )
  expect(screen.getByText('Source identity was not captured.')).toBeTruthy()
  expect(screen.queryByText(/Initial deployment/)).toBeNull()
  expect(screen.queryByText(/Captured source files unchanged/)).toBeNull()

  rerender(
    <InstallDeploymentSourceDiff
      {...props}
      comparison={{
        ...sourceDiff,
        status: 'unsupported',
        before: { id: 'bld-old', digest: 'sha256:old' },
        after: { id: 'bld-new', digest: 'sha256:new' },
      }}
    />
  )
  expect(screen.getByText('sha256:old')).toBeTruthy()
  expect(screen.getByText('sha256:new')).toBeTruthy()
  expect(screen.getByText('Text comparison unsupported')).toBeTruthy()
  expect(document.querySelector('file-tree-container')).toBeNull()
  expect(codeText()).toBe('')
})

test('shows executable-bit changes with identical code, and never treats missing modified content as empty', async () => {
  const mode = sourceDiffContents['deploy.sh']
  const props = {
    comparison: sourceDiff,
    selectedPath: mode.path,
    onSelectPath: () => {},
    fileComparison: { ...sourceDiff, selected_file: mode },
  }
  const { rerender } = renderUI(<InstallDeploymentSourceDiff {...props} />)
  expect(
    screen.getByText(/Executable bit changed: not executable → executable/)
  ).toBeTruthy()
  await waitFor(() => expect(codeText()).toContain('echo'))
  rerender(
    <InstallDeploymentSourceDiff
      {...props}
      fileComparison={{
        ...sourceDiff,
        selected_file: { path: mode.path, after: mode.after },
      }}
    />
  )
  expect(screen.getByText(/File contents unavailable/)).toBeTruthy()
  expect(screen.queryByText('File contents unchanged.')).toBeNull()
  expect(codeText()).toBe('')

  rerender(
    <InstallDeploymentSourceDiff
      {...props}
      comparison={{
        ...sourceDiff,
        files: [
          { path: mode.path, change: 'modified', after_executable: true },
        ],
      }}
    />
  )
  expect(screen.queryByText(/Executable bit changed/)).toBeNull()
  expect(screen.getByText('File contents unchanged.')).toBeTruthy()

  rerender(
    <InstallDeploymentSourceDiff
      {...props}
      comparison={{
        ...sourceDiff,
        files: [{ path: mode.path, change: 'added' }],
      }}
      fileComparison={{
        ...sourceDiff,
        selected_file: { path: mode.path, after: '' },
      }}
    />
  )
  expect(screen.getByText('Empty file added.')).toBeTruthy()
  expect(screen.queryByText('File contents unchanged.')).toBeNull()
})

test('isolates query cache and selected files between install targets and retries manifest errors', async () => {
  const calls: URL[] = []
  let fail = true
  globalThis.fetch = (async (input: string) => {
    const url = new URL(input, 'http://example.test')
    calls.push(url)
    if (fail) return Response.json({ error: 'Unavailable' }, { status: 503 })
    return Response.json({
      ...sourceDiff,
      name: url.pathname.includes('ins-other') ? 'other-api' : 'acme-api',
    })
  }) as typeof fetch
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  clients.push(client)
  const content = (installId: string) => (
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <SurfacesProvider>
          <InstallDeploymentSourceDiffPanel
            target={{ ...target, installId }}
            isVisible
          />
        </SurfacesProvider>
      </MemoryRouter>
    </QueryClientProvider>
  )
  const { rerender } = render(content(target.installId))
  expect(await screen.findByText('Source changes loading failed')).toBeTruthy()
  fail = false
  fireEvent.click(screen.getByRole('button', { name: 'Try again' }))
  expect(await screen.findByText('acme-api')).toBeTruthy()
  await selectFile('logo.png')
  rerender(content('ins-other'))
  expect(await screen.findByText('other-api')).toBeTruthy()
  expect(
    screen.getByText('Select a changed file to view its diff.')
  ).toBeTruthy()
  expect(calls.at(-1)?.pathname).toContain('/installs/ins-other/')
  expect(calls.at(-1)?.searchParams.has('path')).toBe(false)
})
