/** Page context sent with AI chat for page-aware assistance. */
export function buildPageContext(pathname: string, search: Record<string, unknown>): string {
  const parts = [`Current route: ${pathname}`];
  if (typeof search.id === 'string') parts.push(`Entity id: ${search.id}`);
  if (typeof search.traceId === 'string') parts.push(`Trace id: ${search.traceId}`);
  if (typeof search.service === 'string') parts.push(`Service: ${search.service}`);
  if (pathname.startsWith('/incidents/') && pathname !== '/incidents') {
    parts.push(`Incident detail: ${pathname.split('/').pop()}`);
  }
  return parts.join('\n');
}

export function suggestedPromptsForPath(pathname: string): { category: string; question: string }[] {
  if (pathname.startsWith('/incidents/')) {
    return [
      { category: 'Incident', question: 'Summarize this incident and likely root cause' },
      { category: 'Incident', question: 'What logs correlate with this incident?' },
    ];
  }
  if (pathname.startsWith('/logs')) {
    return [
      { category: 'Logs', question: 'Explain the top error patterns in this search' },
      { category: 'Logs', question: 'Which service is driving most errors?' },
    ];
  }
  if (pathname.startsWith('/traces')) {
    return [
      { category: 'Traces', question: 'Which spans are slowest in this trace?' },
      { category: 'Traces', question: 'Compare latency to baseline for this service' },
    ];
  }
  return [
    { category: 'Overview', question: 'What should I investigate first right now?' },
    { category: 'Incidents', question: 'Any open P1 incidents I should know about?' },
  ];
}
