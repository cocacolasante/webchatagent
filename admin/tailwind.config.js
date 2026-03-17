/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#f0eeff',
          100: '#e3deff',
          500: '#6C63FF',
          600: '#5a52e0',
          700: '#4840c0',
        },
      },
    },
  },
  plugins: [],
};
