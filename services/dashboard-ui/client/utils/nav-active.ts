const normalizePath = (path: string) =>
  path.endsWith('/') ? path.slice(0, -1) : path

export function isNavLinkActive(
  basePath: string,
  path: string,
  pathname: string,
  matchPaths?: string[]
): boolean {
  const normalizedPathName = normalizePath(pathname)
  const fullPath = normalizePath(`${basePath}${path}`)

  return (
    fullPath === normalizedPathName ||
    (path !== `/` && normalizedPathName.startsWith(`${fullPath}/`)) ||
    (matchPaths ?? []).some((matchPath) => {
      const fullMatchPath = normalizePath(`${basePath}${matchPath}`)
      return (
        fullMatchPath === normalizedPathName ||
        normalizedPathName.startsWith(`${fullMatchPath}/`)
      )
    })
  )
}
