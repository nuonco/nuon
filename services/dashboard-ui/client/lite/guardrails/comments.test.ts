import { describe, expect, test } from 'bun:test'
import { readdirSync, readFileSync } from 'node:fs'
import { join, relative } from 'node:path'

const LITE_DIR = join(import.meta.dir, '..')

const ALLOWED_DIRECTIVES = [
  /^\/\/\s*eslint-/,
  /^\/\*\s*eslint\b/,
  /^\/\/\s*oxlint-/,
  /^\/\/\s*@ts-/,
  /^\/\/\/\s*<reference/,
  /^\/\/\s*prettier-ignore/,
]

const REGEX_PRECEDERS = new Set([
  '(', ',', '=', ':', '[', '!', '&', '|', '?', '{', '}', ';', '+', '-', '*',
  '%', '~', '^', '<', '>',
])

const REGEX_KEYWORDS = new Set([
  'return', 'typeof', 'case', 'in', 'of', 'new', 'delete', 'void',
  'instanceof', 'yield', 'await',
])

interface IComment {
  line: number
  text: string
}

const skipQuoted = (source: string, start: number, quote: string) => {
  let i = start + 1
  while (i < source.length && source[i] !== quote) {
    i += source[i] === '\\' ? 2 : 1
  }
  return i + 1
}

const skipTemplate = (source: string, start: number) => {
  let i = start + 1
  while (i < source.length) {
    if (source[i] === '\\') {
      i += 2
      continue
    }
    if (source[i] === '`') break
    if (source[i] === '$' && source[i + 1] === '{') {
      let depth = 1
      i += 2
      while (i < source.length && depth > 0) {
        const c = source[i]
        if (c === '{') depth += 1
        if (c === '}') depth -= 1
        if (c === "'" || c === '"') {
          i = skipQuoted(source, i, c) - 1
        } else if (c === '`') {
          i = skipTemplate(source, i) - 1
        }
        i += 1
      }
      continue
    }
    i += 1
  }
  return i + 1
}

const skipRegex = (source: string, start: number) => {
  let i = start + 1
  let inClass = false
  while (i < source.length) {
    if (source[i] === '\\') {
      i += 2
      continue
    }
    if (source[i] === '[') inClass = true
    if (source[i] === ']') inClass = false
    if (source[i] === '\n') break
    if (source[i] === '/' && !inClass) break
    i += 1
  }
  return i + 1
}

export const findComments = (source: string): IComment[] => {
  const comments: IComment[] = []
  const lineAt = (position: number) =>
    source.slice(0, position).split('\n').length

  let i = 0
  let previousChar = ''
  let previousWord = ''

  while (i < source.length) {
    const char = source[i]!
    const nextChar = source[i + 1]

    if (char === '/' && nextChar === '/') {
      const end = source.indexOf('\n', i)
      const stop = end === -1 ? source.length : end
      comments.push({ line: lineAt(i), text: source.slice(i, stop).trim() })
      i = stop
      continue
    }

    if (char === '/' && nextChar === '*') {
      const end = source.indexOf('*/', i + 2)
      const stop = end === -1 ? source.length : end + 2
      comments.push({
        line: lineAt(i),
        text: source.slice(i, stop).split('\n')[0]!.trim(),
      })
      i = stop
      continue
    }

    if (char === "'" || char === '"') {
      i = skipQuoted(source, i, char)
      previousChar = 'x'
      previousWord = ''
      continue
    }

    if (char === '`') {
      i = skipTemplate(source, i)
      previousChar = 'x'
      previousWord = ''
      continue
    }

    if (
      char === '/' &&
      (previousChar === '' ||
        REGEX_PRECEDERS.has(previousChar) ||
        REGEX_KEYWORDS.has(previousWord))
    ) {
      i = skipRegex(source, i)
      previousChar = 'x'
      previousWord = ''
      continue
    }

    if (/\s/.test(char)) {
      i += 1
      continue
    }

    if (/[A-Za-z_$]/.test(char)) {
      let word = ''
      while (i < source.length && /[\w$]/.test(source[i]!)) {
        word += source[i]
        i += 1
      }
      previousWord = word
      previousChar = 'x'
      continue
    }

    previousChar = char
    previousWord = ''
    i += 1
  }

  return comments
}

const walk = (dir: string): string[] =>
  readdirSync(dir, { withFileTypes: true }).flatMap((entry) =>
    entry.isDirectory()
      ? walk(join(dir, entry.name))
      : [join(dir, entry.name)]
  )

const sourceFiles = walk(LITE_DIR)
  .filter((file) => /\.tsx?$/.test(file))
  .sort()

describe('the comment scanner', () => {
  const cases: Array<[string, number]> = [
    ["const a = 'https://x.co'", 0],
    ['const a = "http://x.co"', 0],
    ['const t = `a//b ${x} c`', 0],
    ['const r = /https?:\\/\\//g', 0],
    ['const d = a / b / c', 0],
    ['<Link href="https://docs.nuon.co">docs</Link>', 0],
    ["const s = 'a /* not a comment */ b'", 0],
    ['const a = 1 // trailing', 1],
    ['// leading', 1],
    ['/* block */', 1],
    ['/**\n * jsdoc\n */', 1],
    ['const a = 1 /* inline */ + 2', 1],
    ['const r = /x/ // after regex', 1],
    ['const t = `a` // after template', 1],
    ["const s = 'a' // after string", 1],
    ['return /re/.test(x) // after return regex', 1],
  ]

  for (const [source, expected] of cases) {
    test(`finds ${expected} in ${JSON.stringify(source)}`, () => {
      expect(findComments(source)).toHaveLength(expected)
    })
  }
})

describe('lite source files', () => {
  test('contain no comments', () => {
    const offenders = sourceFiles.flatMap((file) =>
      findComments(readFileSync(file, 'utf8'))
        .filter(
          (comment) =>
            !ALLOWED_DIRECTIVES.some((directive) => directive.test(comment.text))
        )
        .map(
          (comment) =>
            `${relative(LITE_DIR, file)}:${comment.line}  ${comment.text.slice(0, 100)}`
        )
    )

    expect(
      offenders,
      'Lite source files carry no comments. Delete these, and if one is ' +
        'explaining something the code does not, rename the thing or extract ' +
        'it into a named function instead. Tool directives (eslint, oxlint, ' +
        '@ts-, /// <reference>, prettier-ignore) are the only exception. ' +
        'See DEV.md "Comments".'
    ).toEqual([])
  })
})
