export const SCAN_AUTHORIZATION_STATEMENT = 'scan-authorization-v1';

export function scanAuthorization(
  targets: { id: string; value: string }[],
  profile: { nmap_args: string; timeout_seconds: number },
  cron = '',
) {
  return {
    accepted: true,
    statement: SCAN_AUTHORIZATION_STATEMENT,
    targets: targets.map(({ id, value }) => ({ id, value })),
    nmap_args: profile.nmap_args,
    timeout_seconds: profile.timeout_seconds,
    cron_expression: cron,
  };
}
