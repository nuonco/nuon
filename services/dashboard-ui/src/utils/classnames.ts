import classNames from 'classnames'

export function cn(
  ...inputs: (string | undefined | { [key: string]: boolean })[]
) {
  return classNames(...inputs)
}
