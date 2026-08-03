export const toSentenceCase = (str: string = ''): string =>
  str.length ? str.charAt(0).toUpperCase() + str.slice(1).toLowerCase() : ''

export const toTitleCase = (str: string = ''): string =>
  str
    .replace(/[-_]/g, ' ')
    .toLowerCase()
    .replace(/\b(\w)/g, (char) => char.toUpperCase())
    .replace(/\s+/g, ' ')
    .trim()

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

export const kebabToWords = (str: string = ''): string =>
  str ? str.replace(/-/g, ' ') : ''

export const snakeToWords = (str: string = ''): string =>
  str ? str.replace(/_/g, ' ') : ''

export const camelToWords = (str: string = ''): string =>
  str ? str.replace(/([A-Z])/g, ' $1') : ''

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

export const getParentPath = (pathname: string = ''): string => {
  const path =
    pathname.endsWith('/') && pathname.length > 1
      ? pathname.slice(0, -1)
      : pathname

  const lastSlashIndex = path.lastIndexOf('/')

  if (lastSlashIndex <= 0) return '/'

  return path.slice(0, lastSlashIndex) || '/'
}

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

export function getFlagEmoji(countryCode: string = 'us'): string {
  const upperCode = countryCode.toUpperCase()
  return Array.from(upperCode)
    .map((char) => String.fromCodePoint(127397 + char.charCodeAt(0)))
    .join('')
}

export function toOrdinal(n: number): string {
  const j = n % 10
  const k = n % 100

  if (j === 1 && k !== 11) return `${n}st`
  if (j === 2 && k !== 12) return `${n}nd`
  if (j === 3 && k !== 13) return `${n}rd`
  return `${n}th`
}

export function indexToOrdinal(idx: number): string {
  return toOrdinal(idx + 1)
}
