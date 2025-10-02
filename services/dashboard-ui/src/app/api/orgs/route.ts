import { NextResponse, type NextRequest } from 'next/server'
import { getOrgs } from '@/lib'
import { getFetchOpts } from '@/utils'
import { setOrgSessionCookie } from '@/components/org-actions'
import type { TOrg } from '@/types'
import { API_URL } from '@/configs/api'

export const GET = async (request: NextRequest) => {
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined
  const q = searchParams.get('q') || undefined

  const response = await getOrgs({ limit, offset, q })

  return NextResponse.json(response)
}

export const POST = async (request: NextRequest) => {
  try {
    const { name } = await request.json()

    if (!name?.trim()) {
      return NextResponse.json(
        { error: 'Organization name is required' },
        { status: 400 }
      )
    }

    const fetchOpts = await getFetchOpts()

    const res = await fetch(`${API_URL}/v1/orgs`, {
      ...fetchOpts,
      method: 'POST',
      body: JSON.stringify({
        name: name.trim(),
        use_sandbox_mode: true, // Set to true for new orgs as per CLI pattern
      }),
    })


    if (!res.ok) {
      const errorData = await res.json().catch(() => ({}))
      console.error('API error:', errorData, 'Status:', res.status)
      return NextResponse.json(
        {
          error:
            errorData.message ||
            errorData.description ||
            errorData.error ||
            'Unable to create your organization, refresh the page and try again.',
        },
        { status: res.status }
      )
    }

    const newOrg: TOrg = await res.json()

    // Set the org session cookie
    await setOrgSessionCookie(newOrg.id)

    return NextResponse.json({ org: newOrg })
  } catch (error) {
    console.error('Failed to create organization:', error)
    return NextResponse.json(
      {
        error:
          error instanceof Error
            ? error.message
            : 'Failed to create organization. Please try again.',
      },
      { status: 500 }
    )
  }
}
