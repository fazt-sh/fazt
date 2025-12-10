/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
    "./node_modules/preline/dist/*.js", // Add Preline content
  ],
  darkMode: 'class', // Enable class-based dark mode
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: 'rgb(255, 149, 0)',
          dark: 'rgb(230, 134, 0)',
          light: 'rgb(255, 176, 51)',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Consolas', 'Courier New', 'monospace'],
      },
      // Add custom animations
      keyframes: {
        slideIn: {
          '0%': {
            opacity: '0',
            transform: 'translateY(10px)',
          },
          '100%': {
            opacity: '1',
            transform: 'translateY(0)',
          },
        },
      },
      animation: {
        'slide-in': 'slideIn 0.4s ease-out',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'), // Add forms plugin
    // require('preline/plugin'), // Preline plugin - not available yet
  ],
};
