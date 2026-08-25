/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  safelist: [
    {
      pattern: /(bg|text|border|ring)-(emerald|blue|purple|orange|red|amber|teal|cyan)-(50|100|500|600|700|800)/,
      variants: ['hover', 'focus', 'group-hover'],
    },
    'bg-gray-50',
    'bg-white',
    'text-gray-500',
    'text-gray-600',
    'text-gray-700',
    'text-gray-900',
    'border-gray-100',
    'border-gray-200',
    'border-gray-300',
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
      },
      animation: {
        'fade-in': 'fadeIn 0.4s ease-out',
        'slide-up': 'slideUp 0.4s ease-out',
        'slide-up-delay': 'slideUp 0.6s ease-out 0.15s both',
        'fade-in-delay': 'fadeIn 0.6s ease-out 0.3s both',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(20px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
      },
    },
  },
  plugins: [],
}
