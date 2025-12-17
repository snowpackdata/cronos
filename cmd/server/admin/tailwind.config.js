/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/style.css",
    "./src/**/*.{vue,js,ts,jsx,tsx,css}",
  ],
  safelist: [
    'bg-sage',
    'bg-sage-dark',
    'bg-sage-light',
    'bg-sage-pale',
    'text-sage',
    'text-sage-dark',
    'text-sage-light',
    'text-sage-pale',
    'hover:bg-sage',
    'hover:bg-sage-dark',
    'hover:bg-sage-light',
    'hover:bg-sage-pale',
    'hover:text-sage',
    'hover:text-sage-dark',
    'hover:text-sage-light',
    'hover:text-sage-pale',
    'focus-visible:outline-sage'
  ],
  theme: {
    extend: {
      colors: {
        'teal': '#226d87',
        'blue': {
          DEFAULT: '#021c51',
          'light-blue': '#ced9e1',
          'dark-blue': '#0a192f',
          500: '#2196f3'
        },
        'pink': '#be5bda',
        'orange': '#ff7849',
        'green': {
          DEFAULT: '#58837e',
          'light-green': '#b5e7a0',
        },
        'sage': {
          DEFAULT: '#58837e',
          dark: '#476b67',
          light: '#76a19c',    
          pale: '#e6efee'
        },
        'green-sage': {
          DEFAULT: '#58837e',
          dark: '#476b67',
          light: '#76a19c',    
          pale: '#e6efee'
        },
        'gray': {
          DEFAULT: '#d6d7d9',
          'gray-light': '#eeeff1',
          'gray-dark': '#22262a',
          100: '#f3f4f6',
          400: '#9ca3af',
          500: '#6b7280',
          800: '#1F2937',
          900: '#111827'
        },
        'neutral': {
          DEFAULT: '#f4f5f7',
          50: '#f9fafb',
          600: '#6b7280',
          950: '#1a202c'
        },
        'white': '#ffffff',
        'yellow' : '#f7d070',
        'black': '#000000',
        'red': {
          DEFAULT: '#9f0d0d',
          600: '#dc2626',
          700: '#b91c1c',
        },
        'sky': {
          600: '#0284c7',
          700: '#0369a1',
          800: '#075985',
          900: '#0c4a6e',
          950: '#082f49'
        },
        'green-success': '#47d67c',
        'green-success-light': '#e5f3ec',
        'green-success-dark': '#186335',
        'red-error': '#f44336',
        'red-error-light': '#ffebee',
        'red-error-dark': '#b71c1c',
        'whitesmoke': '#f5f5f5',
      },
      boxShadow: {
        'sidebar': '0 4px 12px 0 rgba(0, 0, 0, 0.15)',
        'navbar': '0 2px 4px 0 rgba(0, 0, 0, 0.1)',
        'card': '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
        'card-hover': '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)'
      },
      borderRadius: {
        'xl': '1rem'
      },
      fontFamily: {
        sans: ['Montserrat', 'sans-serif'],
        serif: ['Georgia', 'serif'],
      },
    },
  },

  plugins: [
    require('@tailwindcss/forms'),
  ]
}

