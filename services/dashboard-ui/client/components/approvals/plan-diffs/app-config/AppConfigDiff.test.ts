import { describe, expect, test } from 'bun:test'
import { formatEmbeddedTomlContent } from './AppConfigDiff'

describe('formatEmbeddedTomlContent', () => {
  test('expands embedded contents into a multiline literal', () => {
    expect(
      formatEmbeddedTomlContent(
        `[[var_file]]\ncontents = 'region = "us-west-2"\\nenabled = true'`
      )
    ).toBe(
      `[[var_file]]\ncontents = '''\nregion = "us-west-2"\nenabled = true\n'''`
    )
  })

  test('preserves ordinary strings and single-line contents', () => {
    const source = `repo = 'acme/example'\ncontents = 'enabled = true'`

    expect(formatEmbeddedTomlContent(source)).toBe(source)
  })

  test('preserves assignment indentation', () => {
    expect(formatEmbeddedTomlContent(`  inline_contents = "echo one\\necho two"`))
      .toBe(`  inline_contents = """\necho one\necho two\n  """`)
  })
})
