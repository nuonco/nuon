'use server'

import { NextResponse } from 'next/server'
import { VERSION } from "@/configs/app"

export const GET = async () => {
  return NextResponse.json({ version: VERSION })
}
