import { FeatureCard } from '@/components/FeatureCard'
import { labFeatures } from '@/data/features'

export const FeatureGrid = () => {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      {labFeatures.map((feature) => (
        <FeatureCard key={feature.id} feature={feature} />
      ))}
    </div>
  )
}
