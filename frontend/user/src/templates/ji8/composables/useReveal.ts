import { nextTick, onBeforeUnmount, type Ref } from 'vue'

/**
 * 滚动进入视口时给元素加 `is-in`，配合 ji8 样式里的 `.j8-reveal` 做依次上浮渐显。
 *
 * 只有 JS 接管后才先藏起来（`is-pending`），所以不支持 IntersectionObserver 或
 * 用户开了「减少动态效果」时内容照常直接显示。同一批进入视口的元素按顺序错开 60ms。
 */
export function useReveal(root: Ref<HTMLElement | null>, selector: string) {
  let observer: IntersectionObserver | null = null
  const disabled =
    typeof window === 'undefined' ||
    typeof IntersectionObserver === 'undefined' ||
    window.matchMedia?.('(prefers-reduced-motion: reduce)').matches

  const ensureObserver = () => {
    observer ??= new IntersectionObserver(
      (entries) => {
        entries
          .filter((entry) => entry.isIntersecting)
          .forEach((entry, index) => {
            const el = entry.target as HTMLElement
            el.style.setProperty('--d', `${Math.min(index, 8) * 60}ms`)
            el.classList.remove('is-pending')
            el.classList.add('is-in')
            observer?.unobserve(el)
          })
      },
      { rootMargin: '0px 0px -5% 0px', threshold: 0.06 },
    )
    return observer
  }

  const scan = async () => {
    await nextTick()
    const el = root.value
    if (disabled || !el) return
    const io = ensureObserver()
    el.querySelectorAll<HTMLElement>(selector).forEach((node) => {
      if (node.dataset.reveal) return
      node.dataset.reveal = '1'
      node.classList.add('is-pending')
      io.observe(node)
    })
  }

  onBeforeUnmount(() => {
    observer?.disconnect()
    observer = null
  })

  return { scan }
}
