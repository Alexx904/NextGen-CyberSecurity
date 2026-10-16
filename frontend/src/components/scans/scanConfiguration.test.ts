import { describe, expect, it } from 'vitest';
import { SCAN_PRESETS, canonicalArgs, commandError, isIntrusive, PROFILE_DURATION_ESTIMATES, profileDurationLabel } from './scanConfiguration';

describe('Nmap configuration contract', () => {
  it.each(SCAN_PRESETS)('accepts preset $id and describes explicit options', (preset) => {
    expect(commandError(preset.args)).toBeNull();
    expect(preset.args).not.toContain('-sn');
    const range = PROFILE_DURATION_ESTIMATES[preset.id];
    expect(range[0]).toBeGreaterThan(0); expect(range[1]).toBeGreaterThanOrEqual(range[0]);
    expect(range[1]).toBeLessThanOrEqual(preset.timeout);
    expect(profileDurationLabel(preset.id)).not.toBe('Non disponibile');
  });
  it('normalizes spelling without reordering flags', () => {
    expect(canonicalArgs('nmap -sV -p 80,443 --script vuln')).toBe('-sV -p80,443 --script=vuln');
  });
  it.each(['--script=vuln', '--script vuln', '--script vulners', '-p-', '-p1-65535', '-p 1-65535', '-T4', '-T5', '--version-all', '-sA', '--min-rate 1000'])('requires acknowledgement for %s', (args) => {
    expect(isIntrusive(args)).toBe(true);
  });
  it('allows mixed TCP and UDP port lists', () => {
    expect(commandError('-sS -sU -p T:80,443,U:53,161 -T3')).toBeNull();
  });
});
