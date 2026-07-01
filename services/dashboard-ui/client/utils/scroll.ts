type ScrollBlock = 'start' | 'nearest'

interface IScrollIntoView {
  block?: ScrollBlock
  behavior?: ScrollBehavior
}

const isScrollableY = (el: HTMLElement): boolean => {
  if (el.scrollHeight <= el.clientHeight) return false
  const overflowY = getComputedStyle(el).overflowY
  return overflowY === 'auto' || overflowY === 'scroll'
}

const scrollWithin = (
  el: HTMLElement,
  container: HTMLElement,
  block: ScrollBlock,
  behavior: ScrollBehavior
) => {
  const containerRect = container.getBoundingClientRect()
  const elRect = el.getBoundingClientRect()
  const marginTop = parseFloat(getComputedStyle(el).scrollMarginTop) || 0

  let delta = 0
  if (block === 'start') {
    delta = elRect.top - containerRect.top - marginTop
  } else if (elRect.top < containerRect.top + marginTop) {
    delta = elRect.top - containerRect.top - marginTop
  } else if (elRect.bottom > containerRect.bottom) {
    delta = elRect.bottom - containerRect.bottom
  }

  if (delta !== 0) {
    container.scrollTo({ top: container.scrollTop + delta, behavior })
  }
}

// Native Element.scrollIntoView() scrolls every scroll-container ancestor —
// including overflow:hidden wrappers, which are programmatically scrollable but
// have no scrollbar. In the nested layout that leaves the header shifted out of
// view until a reflow. This scrolls only genuine (auto/scroll) ancestors.
export const scrollElementIntoView = (
  el: HTMLElement | null,
  { block = 'nearest', behavior = 'smooth' }: IScrollIntoView = {}
) => {
  if (!el) return
  let parent = el.parentElement
  while (parent && parent !== document.body) {
    if (isScrollableY(parent)) {
      scrollWithin(el, parent, block, behavior)
    }
    parent = parent.parentElement
  }
}
