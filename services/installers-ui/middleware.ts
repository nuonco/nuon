/*
 * Middleware
 *
 * This project is now able to support two modes.
 *
 *   1. A single organization installer.
 *   2. A multi-tenant installer.
 *
 * In multi-tenant mode, the subdomain is the org slug. It is required
 * and is used to rewrite the request so the user lands in the `/:org-slug`
 * route upon visiting the homepage.
 *
 *   > acme.installer.nuon.co                    -> /:org-slug
 *   > acme.installer.nuon.co/installer-slug     -> /:org-slug/:installer-slug
 *   > acme.installer.nuon.co/installer-slug/:id -> /:org-slug/:installer-slug/:installer-id
 *
 * In single organization mode, there is no need for a subdomain.
 * The organization is configured in the env vars.
 *
 */
import { NextResponse } from 'next/server'

export async function getOrgByName(orgSlug): Promise<Record<string, any>> {
  console.debug(`[getOrgByName] orgSlug=${orgSlug}`)
  const res = await fetch(
    `${process.env.NUON_ADMIN_API_URL}/v1/orgs/admin-get?name=${orgSlug}`,
    {
      cache: 'no-store',
    }
  )
  console.debug(`[getOrgByName] status=${res.status} ok=${res.ok}`)

  return res.json()
}

enum Mode {
  Single = 'single',
  Multi = 'multi',
}
const DO_NOT_TOUCH = ['api', '_next', 'favicon.ico']
export const getTenancyMode = () => {
  if (process.env.NUON_ORG_ID && process.env.NUON_INSTALLER_ID) {
    return Mode.Single
  }
  return Mode.Multi
}

const singleTenantMiddleware = (req) => {
  // get the hostname
  let hostname = req.headers.get('host')
  if (hostname.indexOf(':') >= 1) {
    hostname = hostname.split(':')[0]
  }

  // prepare our additions to the request
  const config = {
    headers: { 'X-Nuon-Org-Id': process.env.NUON_ORG_ID },
  }

  // if we are fetching the API or static files: we leave the request as-is but add our header
  let requestedUrl = new URL(req.url) // so we can get pathname (e.g.): "/1/2/3"
  let basePath = requestedUrl.pathname.split('/')[1]
  if (DO_NOT_TOUCH.includes(basePath)) {
    return NextResponse.next(config)
  }

  // if we are reaching for a subdomain on a single-tenant app, redirect to the NEXT_PUBLIC_WEBAPP_URL
  // a single-tenant app should not worry about subdomains
  console.log(
    `[${hostname === process.env.NEXT_PUBLIC_WEBAPP_URL}] hostname: ${hostname} public_url: ${process.env.NEXT_PUBLIC_WEBAPP_URL}`
  )
  if (hostname !== process.env.NEXT_PUBLIC_WEBAPP_URL) {
    // this is a rough edge for development
    return NextResponse.redirect(
      `https://${process.env.NEXT_PUBLIC_WEBAPP_URL}`
    )
  }

  // we redirect to /:org-slug
  // NOTE:  `org` here is a dummy slug.
  let url = new URL(`/org-slug${req.nextUrl.pathname}`, req.url)
  console.log(`${url}`)
  return NextResponse.rewrite(url, config)
}

const multiTenantMiddleware = async (req) => {
  // get the hostname
  let hostname = req.headers.get('host')
  if (hostname.indexOf(':') >= 1) {
    hostname = hostname.split(':')[0]
  }

  // if the hostname is the root of the webapp, we redirect to the homepage
  // the homepage shouddirect users to the right place
  if (hostname === process.env.NEXT_PUBLIC_WEBAPP_URL) {
    return NextResponse.next()
  }

  const subdomain = hostname.match(/^([^.]+)\./)?.[1]
  const environments = ['local', 'dev', 'stage', 'prod']

  // if there is no subdomain but we've reached an environment subdomain,
  // redirect to the homepage
  if (environments.includes(subdomain)) {
    return NextResponse.next()
  }

  //
  // finally: if we are on still here, the subdomain is a org slug.
  //

  // fetch the org from the admin api and the headers
  let org = await getOrgByName(subdomain) // requires the endpoint to support this
  console.log(org)
  let nuon_org_id = org.id
  const config = {
    headers: { 'X-Nuon-Org-Name': subdomain, 'X-Nuon-Org-Id': nuon_org_id },
  }

  // if we are fetching the API or static files: we leave the request as-is but add our header
  let requestedUrl = new URL(req.url) // so we can get pathname (e.g.): "/1/2/3"
  let basePath = requestedUrl.pathname.split('/')[1]
  if (DO_NOT_TOUCH.includes(basePath)) {
    return NextResponse.next(config)
  }
  // return NextResponse.next();

  // we rewrite the url from 'slug.installers.nuon.dev` to `installers.nuon.dev/slug`
  const url = new URL(`/${subdomain}${req.nextUrl.pathname}`, req.url)
  return NextResponse.rewrite(url, config)
}

export function middleware(req) {
  if (getTenancyMode() === Mode.Single) {
    return singleTenantMiddleware(req)
  }
  return multiTenantMiddleware(req)
}
