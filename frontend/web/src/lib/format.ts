
/**
 * Format an ISO timestamp as HH:MM in UTC.
 * Pinned to en-GB + UTC so server and client render identical strings.
 */
export function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString('en-GB', {
    hour: '2-digit',
    minute: '2-digit',
    timeZone: 'UTC',
  });
}

/**
 * Format an ISO timestamp as "DD Mon" in UTC.
 */
export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    timeZone: 'UTC',
  });
}