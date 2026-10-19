import { afterEach } from 'vitest';
import { cleanup } from '@testing-library/react';

// Pulizia automatica del DOM e dei componenti React montati al termine di ciascun test
afterEach(() => {
  cleanup();
});
