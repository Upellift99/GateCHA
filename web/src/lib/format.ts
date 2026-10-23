// A key's domain setting is a free list (newlines, commas or spaces).
export function domainList(domain: string): string[] {
  return domain.split(/[\s,]+/).filter(Boolean)
}
