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

export const removeSnakeCase = (str?: string) => str ? str?.replace(/_/g, ' ') : "unknown"

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
    case "install_deploys":
      triggeredBy = "Deploy hook"
      break;
    case "install_sandbox_runs":
      triggeredBy = "Sandbox hook"
      break;
    case "install_action_workflow_manual_triggers":
      triggeredBy = "Manual run"
      break;
    default:
      triggeredBy = "Cron"      
  }

  return triggeredBy
}
