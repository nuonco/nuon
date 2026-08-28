export default {
  title: 'Common/AnimatedHeight',
}

import { useState } from 'react'
import { AnimatedHeight } from './AnimatedHeight'
import { Button } from './Button'
import { Card } from './Card'
import { Text } from './Text'

export const Default = () => {
  const [tall, setTall] = useState(false)
  return (
    <div className="flex flex-col gap-4">
      <Button
        variant="secondary"
        className="w-fit"
        onClick={() => setTall(!tall)}
      >
        Toggle content height
      </Button>
      <Card>
        <AnimatedHeight>
          <div className={tall ? 'h-96' : 'h-24'}>
            <Text>
              This content is {tall ? 'tall (24rem)' : 'short (6rem)'} — the
              wrapper animates between the two heights.
            </Text>
          </div>
        </AnimatedHeight>
      </Card>
    </div>
  )
}
