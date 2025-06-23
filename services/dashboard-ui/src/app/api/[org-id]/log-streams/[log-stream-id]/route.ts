import { NextRequest, NextResponse } from 'next/server'
import { getLogStream } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'log-stream-id'>
) => {
  const { ['org-id']: orgId, ['log-stream-id']: logStreamId } = await params

  let logStream = {}
  try {
    logStream = await getLogStream({ orgId, logStreamId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(logStream)
}
