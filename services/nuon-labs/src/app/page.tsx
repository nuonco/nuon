'use client'

import { Header } from '@/components/Header'
import { Footer } from '@/components/Footer'
import { FeatureGrid } from '@/components/FeatureGrid'
import { ThreeBackground } from '@/components/ThreeBackground'

export default function Home() {
  return (
    <main className="relative min-h-screen flex flex-col">
      <ThreeBackground />
      <Header />

      <div className="flex-1 flex flex-col justify-end px-6 py-12">
        <div className="max-w-6xl mx-auto w-full">
          <FeatureGrid />
        </div>
      </div>

      <Footer />
    </main>
  )
}
