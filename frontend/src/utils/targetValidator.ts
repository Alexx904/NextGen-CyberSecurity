export type TargetType = 'IP' | 'CIDR' | 'RANGE' | 'FQDN';

export interface TargetValidationResult {
  isValid: boolean;
  type: TargetType | null;
  error?: string;
  isIPv6?: boolean;
}

const SHELL_METACHARS = [';', '&', '|', '$', '>', '<', '`', '\\', '"', "'", '(', ')', '{', '}', '*', '?', '!'];

export function validateTargetInput(input: string): TargetValidationResult {
  const clean = input.trim();

  if (!clean) {
    return { isValid: false, type: null, error: 'Il target non può essere vuoto' };
  }

  // Prevenzione Flag / Path injection
  if (clean.startsWith('-') || clean.startsWith('/')) {
    return {
      isValid: false,
      type: null,
      error: "Il target non può iniziare con '-' o '/' (prevenzione command injection)",
    };
  }

  // Prevenzione spazi interni
  if (/\s/.test(clean)) {
    return {
      isValid: false,
      type: null,
      error: 'Il target non può contenere spazi',
    };
  }

  // Prevenzione metacaratteri shell
  for (const ch of SHELL_METACHARS) {
    if (clean.includes(ch)) {
      return {
        isValid: false,
        type: null,
        error: `Carattere non consentito '${ch}' rilevato`,
      };
    }
  }

  // 1. Verifica IP (IPv4)
  const ipv4Regex = /^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}$/;
  if (ipv4Regex.test(clean)) {
    return { isValid: true, type: 'IP', isIPv6: false };
  }

  // 1b. Verifica IP (IPv6)
  const ipv6Regex = /^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:))$/;
  if (ipv6Regex.test(clean)) {
    return { isValid: true, type: 'IP', isIPv6: true };
  }

  // 2. Verifica CIDR (IPv4: /1-32, IPv6: /1-128)
  const cidrParts = clean.split('/');
  if (cidrParts.length === 2) {
    const [ipPart, maskStr] = cidrParts;
    const mask = parseInt(maskStr, 10);
    if (!isNaN(mask)) {
      if (ipv4Regex.test(ipPart) && mask >= 1 && mask <= 32) {
        return { isValid: true, type: 'CIDR', isIPv6: false };
      }
      if (ipv6Regex.test(ipPart) && mask >= 1 && mask <= 128) {
        return { isValid: true, type: 'CIDR', isIPv6: true };
      }
    }
  }

  // 3. Verifica IP Range (es: 192.168.1.1-50 o 192.168.1.1-192.168.1.50)
  const rangeRegex = /^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.){3}(25[0-5]|(2[0-4]|1\d|[1-9]|)\d)-(\d{1,3}|((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.){3}(25[0-5]|(2[0-4]|1\d|[1-9]|)\d))$/;
  if (rangeRegex.test(clean)) {
    return { isValid: true, type: 'RANGE', isIPv6: false };
  }

  // 4. Verifica FQDN (Fully Qualified Domain Name)
  const fqdnRegex = /^(?=.{1,253}$)(?!-)[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,63}$/;
  if (fqdnRegex.test(clean)) {
    return { isValid: true, type: 'FQDN', isIPv6: false };
  }

  return {
    isValid: false,
    type: null,
    error: 'Formato bersaglio non valido (atteso IPv4, IPv6, CIDR, Range o FQDN)',
  };
}

