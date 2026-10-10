import { afterEach } from 'vitest';
import { cleanup } from '@testing-library/react';

// Pulizia automatica del DOM e dei componenti React montati al termine di ciascun test
afterEach(() => {
  cleanup();
});

// Soppressione dei warning informativi sulle future flag di migrazione a React Router v7 nei test
const originalWarn = console.warn;
console.warn = (...args: unknown[]) => {
  if (typeof args[0] === 'string' && args[0].includes('React Router Future Flag Warning')) {
    return;
  }
  originalWarn(...args);
};

