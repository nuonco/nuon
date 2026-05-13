type CnInput = string | undefined | false | null | { [key: string]: boolean }

export function cn(...inputs: CnInput[]) {
  const classes: string[] = []
  for (const input of inputs) {
    if (!input) continue
    if (typeof input === 'string') {
      classes.push(input)
    } else {
      for (const key in input) {
        if (input[key]) classes.push(key)
      }
    }
  }
  return classes.join(' ')
}
