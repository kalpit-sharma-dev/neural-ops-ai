/** Theme-aware chart palette — reads CSS variables at render time. */
export function getChartColors(): string[] {
  if (typeof document === 'undefined') {
    return ['#22d3ee', '#a78bfa', '#fbbf24', '#fb7185'];
  }
  const style = getComputedStyle(document.documentElement);
  return ['--chart-1', '--chart-2', '--chart-3', '--chart-4'].map((v) =>
    style.getPropertyValue(v).trim() || '#22d3ee',
  );
}
