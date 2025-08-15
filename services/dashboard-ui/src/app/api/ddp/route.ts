import { NextRequest, NextResponse } from 'next/server'

async function proxyRequest(request: NextRequest) {
  // The `ddforward` query parameter is used to specify the exact endpoint to
  // which the RUM data should be forwarded. The upstream uri to /api/v2/rum?ddsource=...
  // is URL encoded in the ddforward query parameter.
  //
  // Docs: https://docs.datadoghq.com/real_user_monitoring/guide/proxy-rum-data/?lib_src=npm&rum_browser_sdk_version=gte_5_4_0&site=us5

  const ddUrl = 'https://browser-intake-us5-datadoghq.com'
  const ddForward = request.nextUrl.searchParams.get('ddforward')
  const upstreamURL = `${ddUrl}${ddForward}`
  const requestHeaders = buildRequestHeaders(request)
  const fetchInit: RequestInit = {
    method: request.method,
    headers: requestHeaders,
    body: await request.text(),
  }
  const response = await fetch(upstreamURL, fetchInit)
  return new NextResponse(response.body, {
    status: response.status,
    headers: buildResponseHeaders(response.headers),
  })
}

function buildRequestHeaders(request: NextRequest) {
  // These headers shouldn't be proxied upstream
  const skipHeaders = [
    'host',
    'content-length',
    'transfer-encoding',
    'connection',
    'keep-alive',
    'proxy-authorization',
    'te',
    'upgrade',
    'expect',
  ]
  const filteredHeaders = new Headers()
  request.headers.forEach((value, key) => {
    if (!skipHeaders.includes(key.toLowerCase())) {
      filteredHeaders.append(key, value)
    }
  })

  // Add client IP to X-Forwarded-For header
  const clientIP = request.headers.get('x-real-ip')
  if (clientIP) {
    let xForwardedFor = filteredHeaders.get('x-forwarded-for')
    xForwardedFor = xForwardedFor ? `${clientIP}, ${xForwardedFor}` : clientIP
    filteredHeaders.set('x-forwarded-for', xForwardedFor)
  }

  return filteredHeaders
}

function buildResponseHeaders(headers: Headers) {
  // These headers should not be proxied downstream
  const skipHeaders = [
    'connection',
    'keep-alive',
    'proxy-authenticate',
    'transfer-encoding',
    'upgrade',
    'trailer',
  ]
  const filteredHeaders = new Headers()
  headers.forEach((value, key) => {
    if (!skipHeaders.includes(key.toLowerCase())) {
      filteredHeaders.append(key, value)
    }
  })
  return filteredHeaders
}

export async function GET(request: NextRequest) {
  return proxyRequest(request)
}
export async function POST(request: NextRequest) {
  return proxyRequest(request)
}
