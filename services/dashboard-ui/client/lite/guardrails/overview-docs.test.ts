import { describe, expect, test } from 'bun:test'
import { existsSync, readdirSync, readFileSync } from 'node:fs'
import { basename, dirname, join, relative } from 'node:path'

const COMPONENTS_DIR = join(import.meta.dir, '..', 'components')
const SKIP_DIRS = new Set(['__stories__', '__fixtures__'])

const INTERNAL_COMPONENTS = new Set([
  'organisms/surfaces/SurfaceOverlay',
  'organisms/surfaces/SurfaceTransition',
  'organisms/toasts/ToastStack',
])

const KNOWN_GAPS: Record<string, string[]> = {}

const walk = (dir: string): string[] =>
  readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    if (entry.isDirectory()) {
      return SKIP_DIRS.has(entry.name) ? [] : walk(join(dir, entry.name))
    }
    return entry.isFile() ? [join(dir, entry.name)] : []
  })

const allFiles = walk(COMPONENTS_DIR)
const storyFiles = allFiles.filter((file) => file.endsWith('.stories.tsx')).sort()
const keyOf = (file: string) =>
  relative(COMPONENTS_DIR, file).replace(/\.tsx$/, '')

const balancedBlock = (source: string, opener: string) => {
  const start = source.indexOf(opener)
  if (start === -1) return null

  let depth = 0
  for (let i = start + opener.length - 1; i < source.length; i += 1) {
    if (source[i] === '[') depth += 1
    if (source[i] === ']') {
      depth -= 1
      if (depth === 0) return source.slice(start + opener.length, i)
    }
  }
  return null
}

const entryNames = (block: string | null) =>
  block ? [...block.matchAll(/\bname:\s*'([^']+)'/g)].map((match) => match[1]!) : []

const quotedCount = (block: string | null) =>
  block ? [...block.matchAll(/'(?:[^'\\]|\\.)*'/g)].length : 0

const interfaceProps = (source: string, interfaceName: string) => {
  const header = new RegExp(`export interface ${interfaceName}\\b[^{]*\\{`).exec(
    source
  )
  if (!header) return null

  const bodyStart = header.index + header[0].length
  let depth = 1
  let bodyEnd = bodyStart
  for (let i = bodyStart; i < source.length; i += 1) {
    if (source[i] === '{') depth += 1
    if (source[i] === '}') {
      depth -= 1
      if (depth === 0) {
        bodyEnd = i
        break
      }
    }
  }

  const props: string[] = []
  let nested = 0
  for (const line of source.slice(bodyStart, bodyEnd).split('\n')) {
    if (nested === 0) {
      const declaration = /^\s*(?:readonly\s+)?([A-Za-z_$][\w$]*)\??\s*:/.exec(line)
      if (declaration) props.push(declaration[1]!)
    }
    nested += (line.match(/[{([]/g) ?? []).length
    nested -= (line.match(/[})\]]/g) ?? []).length
    if (nested < 0) nested = 0
  }

  return [...new Set(props)]
}

const allInterfaceProps = (source: string) =>
  [...source.matchAll(/export interface (I[A-Za-z0-9_]*)\b/g)].flatMap(
    (match) => interfaceProps(source, match[1]!) ?? []
  )

const hasExternalProps = (source: string) =>
  /export interface I[A-Za-z0-9_]*[\s\S]{0,200}?\bextends\b/.test(source) ||
  /^import type \{[^}]*\b[IT][A-Z]/m.test(source)

interface ICoveredComponent {
  key: string
  name: string
  ownProps: string[]
  visibleProps: string[]
  external: boolean
}

const coveredComponents = (
  storyFile: string,
  storySource: string
): ICoveredComponent[] => {
  const dir = dirname(storyFile)
  const base = basename(storyFile, '.stories.tsx')
  const siblings = [
    ...storySource.matchAll(/from '\.\/([A-Za-z0-9_]+)'/g),
  ].map((match) => match[1]!)

  return [base, ...siblings]
    .filter((name, index, names) => {
      if (names.indexOf(name) !== index) return false
      if (!existsSync(join(dir, `${name}.tsx`))) return false
      return name === base || !existsSync(join(dir, `${name}.stories.tsx`))
    })
    .flatMap((name) => {
      const file = join(dir, `${name}.tsx`)
      const source = readFileSync(file, 'utf8')
      const ownProps = interfaceProps(source, `I${name}`)
      if (!ownProps) return []

      return [
        {
          key: keyOf(file),
          name,
          ownProps,
          visibleProps: allInterfaceProps(source),
          external: hasExternalProps(source),
        },
      ]
    })
}

describe('Overview stories', () => {
  test('every component with a props interface is covered by an Overview', () => {
    const covered = new Set(
      storyFiles.flatMap((file) =>
        coveredComponents(file, readFileSync(file, 'utf8')).map(
          (component) => component.key
        )
      )
    )

    const uncovered = allFiles
      .filter(
        (file) =>
          file.endsWith('.tsx') &&
          !file.endsWith('.stories.tsx') &&
          !file.endsWith('.test.tsx') &&
          !file.endsWith('Container.tsx')
      )
      .filter((file) =>
        interfaceProps(readFileSync(file, 'utf8'), `I${basename(file, '.tsx')}`)
      )
      .map(keyOf)
      .filter((key) => !covered.has(key) && !INTERNAL_COMPONENTS.has(key))
      .sort()

    expect(
      uncovered,
      'Every component that exports a props interface needs an Overview — ' +
        'its own stories file, or a sibling one that imports it. See DEV.md ' +
        '"Stories and the Overview requirement".'
    ).toEqual([])
  })

  for (const storyFile of storyFiles) {
    describe(relative(COMPONENTS_DIR, storyFile), () => {
      const source = readFileSync(storyFile, 'utf8')
      const covered = coveredComponents(storyFile, source)
      const documented = entryNames(balancedBlock(source, 'props={['))

      test('Overview is the first export and renders ComponentDocs', () => {
        expect(/^export const ([A-Za-z0-9_]+)/m.exec(source)?.[1]).toBe(
          'Overview'
        )
        expect(source).toContain('<ComponentDocs')
      })

      test('fills summary, use, avoid and rules', () => {
        expect(
          /summary=(?:"[^"]+"|\{)/.test(source),
          'ComponentDocs needs a non-empty summary.'
        ).toBe(true)

        for (const field of ['use', 'avoid', 'rules'] as const) {
          expect(
            quotedCount(balancedBlock(source, `${field}={[`)),
            `ComponentDocs needs at least one ${field} entry. ` +
              'See DEV.md "The bar for each field".'
          ).toBeGreaterThan(0)
        }
      })

      for (const component of covered) {
        const allowed = KNOWN_GAPS[component.key] ?? []

        test(`documents every prop of ${component.name}`, () => {
          const missing = component.ownProps.filter(
            (prop) => !documented.includes(prop) && !allowed.includes(prop)
          )

          expect(
            missing,
            `${component.name} has props with no entry in its Overview. ` +
              'Add them to props={[...]} — changing a component\'s props ' +
              'means changing its Overview in the same change.'
          ).toEqual([])
        })

        if (allowed.length) {
          test(`${component.name} KNOWN_GAPS entries are still accurate`, () => {
            expect(
              allowed.filter((prop) => documented.includes(prop)),
              `${component.name} now documents these. Remove them from ` +
                'KNOWN_GAPS so the gap cannot reopen.'
            ).toEqual([])

            expect(
              allowed.filter((prop) => !component.ownProps.includes(prop)),
              `${component.name} no longer has these props. Remove them ` +
                'from KNOWN_GAPS.'
            ).toEqual([])
          })
        }
      }

      if (covered.length && !covered.some((component) => component.external)) {
        test('documents no props that no longer exist', () => {
          const real = new Set(
            covered.flatMap((component) => component.visibleProps)
          )

          expect(
            documented.filter((prop) => !real.has(prop)),
            'These props are documented but are on no covered component. ' +
              'They were renamed or removed and the Overview was not updated.'
          ).toEqual([])
        })
      }
    })
  }
})
