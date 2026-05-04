/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        bg: '#0f1117',
        surface: '#1a1d27',
        border: '#2a2d37',
        text: '#d1d4dc',
        'text-secondary': '#787b86',
        green: '#00bcd4',
        red: '#ef5350',
        blue: '#42a5f5',
        yellow: '#ffca28',
      }
    },
  },
  plugins: [],
}
