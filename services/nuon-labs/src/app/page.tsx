'use client'

import { Footer } from '@/components/Footer'
import { Hero } from '@/components/Hero'
import PaperBackground from '@/components/PaperBackground'

export default function Home() {
  return (
    <main className="relative min-h-screen flex flex-col">
      <PaperBackground />

      <Hero />

      <Footer />
    </main>
  )
}
