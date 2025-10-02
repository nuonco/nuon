/**
 * String utility functions.
 * All helpers are pure and safe for general use.
 */

/**
 * Transform a string to sentence case.
 * Example: "hello world" -> "Hello world"
 */
export const toSentenceCase = (str: string = ''): string =>
  str.length ? str.charAt(0).toUpperCase() + str.slice(1).toLowerCase() : ''

/**
 * Transform a string to title case.
 * Example: "hello_world-foo" -> "Hello World Foo"
 */
export const toTitleCase = (str: string = ''): string =>
  str
    .replace(/[-_]/g, ' ')
    .toLowerCase()
    .replace(/\b(\w)/g, (char) => char.toUpperCase())
    .replace(/\s+/g, ' ')
    .trim()

/**
 * Get the initials from a string.
 * Examples:
 *   "John Doe" -> "JD"
 *   "alice" -> "A"
 */
export const getInitials = (str: string = ''): string => {
  if (!str) return ''
  const words = str
    .replace(/[_\-]+/g, ' ')
    .split(' ')
    .filter(Boolean)

  if (words.length === 1) {
    return words[0].charAt(0).toUpperCase()
  }
  return (
    (words[0]?.charAt(0) || '') + (words[words.length - 1]?.charAt(0) || '')
  ).toUpperCase()
}

/**
 * Convert kebab-case to words (spaces).
 * Example: "foo-bar-baz" -> "foo bar baz"
 */
export const kebabToWords = (str: string = ''): string =>
  str ? str.replace(/-/g, ' ') : ''

/**
 * Convert snake_case to words (spaces).
 * Example: "foo_bar_baz" -> "foo bar baz"
 */
export const snakeToWords = (str: string = ''): string =>
  str ? str.replace(/_/g, ' ') : ''

/**
 * Convert camelCase to words (spaces).
 * Example: "exampleString" -> "example String"
 */
export const camelToWords = (str: string = ''): string =>
  str ? str.replace(/([A-Z])/g, ' $1') : ''

/**
 * Slugify a string (URL-safe, lowercase, removes diacritics).
 * Example: "Hello World!" -> "hello-world"
 */
export const slugify = (str: string = ''): string =>
  str
    .toString()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-+|-+$/g, '')

/**
 * Get the parent path of a given path string.
 * Example: "/foo/bar/baz/" -> "/foo/bar"
 *          "/foo" -> "/"
 */
export const getParentPath = (pathname: string = ''): string => {
  const path =
    pathname.endsWith('/') && pathname.length > 1
      ? pathname.slice(0, -1)
      : pathname

  const lastSlashIndex = path.lastIndexOf('/')

  if (lastSlashIndex <= 0) return '/'

  return path.slice(0, lastSlashIndex) || '/'
}

/**
 * Format a byte value as a human-readable string (GB, MB, KB, Bytes).
 * Example: 1048576 -> "1.00 MB"
 */
export const formatBytes = (bytes: number): string => {
  const KB = 1024
  const MB = KB * KB
  const GB = KB * KB * KB

  return bytes >= GB
    ? `${(bytes / GB).toFixed(2)} GB`
    : bytes >= MB
      ? `${(bytes / MB).toFixed(2)} MB`
      : bytes >= KB
        ? `${(bytes / KB).toFixed(2)} KB`
        : `${bytes} Bytes`
}

/**
 * Get the flag emoji for the given country code.
 * Example: "us" -> 🇺🇸
 */
export function getFlagEmoji(countryCode: string = 'us'): string {
  const upperCode = countryCode.toUpperCase()
  return Array.from(upperCode)
    .map((char) => String.fromCodePoint(127397 + char.charCodeAt(0)))
    .join('')
}
