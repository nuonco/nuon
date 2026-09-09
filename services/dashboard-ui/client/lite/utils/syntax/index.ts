import {
  registerCustomCSSVariableTheme,
  registerCustomLanguage,
  type LanguageRegistration,
} from '@pierre/diffs'

export const LITE_SYNTAX_THEME = 'lite'

const TOKEN_DEFAULTS: Record<string, string> = {
  foreground: 'var(--syntax-plain)',
  background: 'var(--code-bg)',
  'token-comment': 'var(--syntax-comment)',
  'token-string': 'var(--syntax-string)',
  'token-string-expression': 'var(--syntax-string-expression)',
  'token-constant': 'var(--syntax-constant)',
  'token-keyword': 'var(--syntax-keyword)',
  'token-function': 'var(--syntax-function)',
  'token-parameter': 'var(--syntax-parameter)',
  'token-punctuation': 'var(--syntax-punctuation)',
  'token-link': 'var(--syntax-link)',
}

const ALIASES: Record<string, string> = {
  sh: 'shellscript',
  shell: 'shellscript',
  bash: 'shellscript',
  zsh: 'shellscript',
  yml: 'yaml',
  tf: 'terraform',
  tfvars: 'terraform',
  md: 'markdown',
  mdx: 'markdown',
  dockerfile: 'docker',
  mmd: 'mermaid',
  txt: 'text',
  plain: 'text',
}

export const SUPPORTED_LANGUAGES = [
  'shellscript',
  'json',
  'yaml',
  'hcl',
  'terraform',
  'toml',
  'markdown',
  'docker',
  'mermaid',
  'rego',
  'text',
] as const

export type TSyntaxLanguage = (typeof SUPPORTED_LANGUAGES)[number]

const warned = new Set<string>()

export const resolveLanguage = (language?: string): TSyntaxLanguage => {
  if (!language) return 'text'

  const normalized = language.toLowerCase()
  const resolved = ALIASES[normalized] ?? normalized

  if (!SUPPORTED_LANGUAGES.includes(resolved as TSyntaxLanguage)) {
    if (process.env.NODE_ENV === 'development' && !warned.has(normalized)) {
      warned.add(normalized)
      console.warn(
        `[lite] Unknown code language "${language}", rendering as plain text. ` +
          `Add it to SUPPORTED_LANGUAGES in client/lite/syntax if it should highlight.`
      )
    }
    return 'text'
  }

  return resolved as TSyntaxLanguage
}

let registered = false

export const registerSyntax = () => {
  if (registered) return
  registered = true

  registerCustomCSSVariableTheme(LITE_SYNTAX_THEME, TOKEN_DEFAULTS)
  registerCustomLanguage(
    'rego',
    () =>
      import('./rego.tmLanguage.json').then((module) => ({
        default: [module.default as unknown as LanguageRegistration],
      })),
    ['rego']
  )
}
