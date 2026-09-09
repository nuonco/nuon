import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, waitFor } from '@testing-library/react'
import {
  composePageTitle,
  PageTitleProvider,
} from '../providers/page-title-provider'
import { usePageTitle } from './use-page-title'

afterEach(cleanup)

const Page = ({ segments }: { segments: (string | undefined)[] }) => {
  usePageTitle(...segments)
  return null
}

const renderPage = (segments: (string | undefined)[]) =>
  render(
    <PageTitleProvider>
      <Page segments={segments} />
    </PageTitleProvider>
  )

describe('composePageTitle', () => {
  test('joins segments most specific first', () => {
    expect(composePageTitle(['Components', 'acme-production'])).toBe(
      'Components | acme-production'
    )
  })

  test('drops unresolved and empty segments', () => {
    expect(composePageTitle(['Components', undefined])).toBe('Components')
    expect(composePageTitle([undefined, 'acme-production'])).toBe(
      'acme-production'
    )
    expect(composePageTitle(['Components', ''])).toBe('Components')
    expect(composePageTitle([null, false, 'Apps'])).toBe('Apps')
  })
})

describe('usePageTitle', () => {
  test('appends the app name to the route title', async () => {
    renderPage(['Installs'])

    await waitFor(() => expect(document.title).toBe('Installs | Nuon'))
  })

  test('leaves an owning entity out until it resolves', async () => {
    const view = renderPage(['Components', undefined])

    await waitFor(() => expect(document.title).toBe('Components | Nuon'))

    view.rerender(
      <PageTitleProvider>
        <Page segments={['Components', 'acme-production']} />
      </PageTitleProvider>
    )

    await waitFor(() =>
      expect(document.title).toBe('Components | acme-production | Nuon')
    )
  })

  test('falls back to the app name when the owning route unmounts', async () => {
    const view = renderPage(['Installs'])

    await waitFor(() => expect(document.title).toBe('Installs | Nuon'))

    view.rerender(<PageTitleProvider>{null}</PageTitleProvider>)

    await waitFor(() => expect(document.title).toBe('Nuon'))
  })
})
