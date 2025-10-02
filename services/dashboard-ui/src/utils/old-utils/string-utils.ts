export const sentanceCase = (s = '') => s.charAt(0).toUpperCase() + s.slice(1)
export const titleCase = (s = '') =>
  s.replace(/^_*(.)|_+(.)/g, (s, c, d) => (c ? c.toUpperCase() : ' ' + d))
export const initialsFromString = (s = '') => {
  // @ts-ignore: IterableIterator<RegExpExecArray>
  const initials = [...s.matchAll(new RegExp(/(\p{L}{1})\p{L}+/, 'gu'))] || []

  return (
    (initials.shift()?.[1] || '') + (initials.pop()?.[1] || '')
  ).toUpperCase()
}

export const removeKebabCase = (str: string): string => str ? str.replace(/-/g, ' ') : 'unknown';

export const removeSnakeCase = (str?: string) =>
  str ? str?.replace(/_/g, ' ') : 'unknown'

export const slugifyString = (str: string) =>
  str
    .toString()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')

export function removeLastPathSegment(pathname: string) {
  let path = pathname.endsWith('/') && pathname.length > 1
    ? pathname.slice(0, -1)
    : pathname;

  const lastSlashIndex = path.lastIndexOf('/');

  if (lastSlashIndex <= 0) return '/';

  return path.slice(0, lastSlashIndex) || '/';
}

export const sizeToMbOrGB = (bytes: number): string => {
  const KB = 1024
  const MB = 1024 ** 2 // 1 MB = 1024 * 1024 bytes
  const GB = 1024 ** 3 // 1 GB = 1024 * 1024 * 1024 bytes

  return bytes >= GB
    ? `${(bytes / GB).toFixed(2)} GB`
    : bytes >= MB
      ? `${(bytes / MB).toFixed(2)} MB`
      : bytes >= KB
        ? `${(bytes / KB).toFixed(2)} KB`
        : `${bytes} Bytes`
}

export function getFlagEmoji(countryCode = 'us') {
  const codePoints = countryCode
    .toUpperCase()
    .split('')
    .map((char) => 127397 + char.charCodeAt(0))
  return String.fromCodePoint(...codePoints)
}

export function humandReadableTriggeredBy(triggeredByType: string) {
  let triggeredBy: string

  switch (triggeredByType) {
    case 'install_deploys':
      triggeredBy = 'Deploy hook'
      break
    case 'install_sandbox_runs':
      triggeredBy = 'Sandbox hook'
      break
    case 'pre-component-deploy':
      triggeredBy = 'Pre deploy hook'
      break
    case 'install_action_workflow_manual_triggers':
      triggeredBy = 'Manual run'
      break
    default:
      triggeredBy = 'Cron'
  }

  return triggeredBy
}
