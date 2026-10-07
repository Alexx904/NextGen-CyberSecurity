/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        cyber: {
          dark: '#0a0d14',
          card: '#111726',
          border: '#1e293b',
          accent: '#06b6d4',
          hover: '#0891b2',
        },
        severity: {
          critical: '#dc2626',
          high: '#ea580c',
          medium: '#eab308',
          low: '#3b82f6',
          info: '#64748b',
        }
      }
    },
  },
  plugins: [],
}

