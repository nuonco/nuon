import { type NextRequest, NextResponse } from 'next/server'
import { getAccount } from '@/lib'

export async function GET(_: NextRequest) {
  const response = await getAccount()
  return NextResponse.json(response)
}
