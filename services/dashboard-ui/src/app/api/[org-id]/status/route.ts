import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { API_URL, getFetchOpts } from '@/utils'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId] = req.url.split('/').slice(4, 5)

  let status = {
    orgStatus: {
      status: 'error',
      status_description: 'Failed to get org status',
    },
    orgHealth: {
      status: 'error',
      status_description: 'Failed to get org health',
    },
  }

  try {
    const [org, health] = await Promise.all([
      fetch(`${API_URL}/v1/orgs/current`, await getFetchOpts(orgId)).then((d) =>
        d?.json()
      ),
      fetch(
        `${API_URL}/v1/orgs/current/health-checks`,
        await getFetchOpts(orgId)
      ).then((d) => d?.json()),
    ])

    status = {
      orgStatus: {
        status: org?.status,
        status_description: org?.status_description,
      },
      orgHealth: {
        status: health?.[0]?.status,
        status_description: health?.[0]?.status_description,
      },
    }
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(status)
})
