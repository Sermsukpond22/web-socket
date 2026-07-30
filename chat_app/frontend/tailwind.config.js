/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        instagram: {
          blue: '#0095f6',
          'blue-hover': '#1877f2',
          gray: '#dbdbdb',
          'light-gray': '#fafafa',
          dark: '#262626',
          secondary: '#737373',
          bubble: '#efefef',
          border: '#eef1f4'
        }
      }
    },
  },
  plugins: [],
}
