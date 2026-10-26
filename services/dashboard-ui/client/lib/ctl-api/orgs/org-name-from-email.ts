// Matched on the domain's name label, so country variants (yahoo.co.uk,
// hotmail.fr, gmx.de) count too.
const FREE_EMAIL_NAMES = new Set([
  'gmail',
  'googlemail',
  'yahoo',
  'ymail',
  'rocketmail',
  'outlook',
  'hotmail',
  'live',
  'msn',
  'icloud',
  'me',
  'mac',
  'aol',
  'aim',
  'proton',
  'protonmail',
  'pm',
  'gmx',
  'mail',
  'email',
  'zoho',
  'zohomail',
  'yandex',
  'hey',
  'fastmail',
  'tutanota',
  'tuta',
  'hushmail',
  'mailfence',
  'inbox',
  'qq',
  '163',
  '126',
  'naver',
  'daum',
  'hanmail',
  'rediffmail',
  'web',
  'comcast',
  'verizon',
  'att',
  'sbcglobal',
  'bellsouth',
  'cox',
  'charter',
  'earthlink',
  'btinternet',
  'orange',
  'free',
  'laposte',
  'libero',
  'seznam',
  'wp',
  'onet',
  'interia',
  'rambler',
  'ukr',
  'bigpond',
  'shaw',
  'rogers',
  'sympatico',
  'duck',
  'skiff',
  'example',
])

const SECOND_LEVEL = new Set(['co', 'com', 'org', 'net', 'ac', 'gov', 'edu', 'ltd', 'plc', 'or', 'ne', 'go'])

// The org name for a work email: the domain's name without its suffix, with the
// first letter capitalized (jane@acme.io → "Acme"). Undefined for free email
// providers and anything that does not parse.
export function orgNameFromEmail(email?: string): string | undefined {
  const domain = email?.split('@')[1]?.trim().toLowerCase()
  if (!domain || !email?.includes('@')) return undefined

  const labels = domain.split('.').filter(Boolean)
  if (labels.length < 2) return undefined
  labels.pop()
  const last = labels.at(-1)
  if (labels.length > 1 && last && SECOND_LEVEL.has(last) && domain.split('.').at(-1)!.length === 2) {
    labels.pop()
  }

  const name = labels.at(-1)
  if (!name || FREE_EMAIL_NAMES.has(name) || !/^[a-z0-9-]+$/.test(name)) return undefined
  return name.charAt(0).toUpperCase() + name.slice(1)
}
