export const MOBILE_BREAKPOINT = 768

export function useMobile() {
  const { width } = useWindowSize()
  const isMobile = computed(() => width.value <= MOBILE_BREAKPOINT)

  return {
    isMobile,
  }
}
