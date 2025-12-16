'use client'

import { Footer } from '@/components/Footer'
import { Hero } from '@/components/Hero'
import PaperBackground from '@/components/PaperBackground'
import { CustomCursor } from '@/components/CustomCursor'

export default function Home() {
  return (
    <main className="relative min-h-screen flex flex-col">
      <CustomCursor />
      <PaperBackground />

      <Hero />

      <Footer />
    </main>
  )
}
