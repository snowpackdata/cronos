/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/**/*.{html, gohtml}", "./templates/*.{html, gohtml}", "./assets/**/*.js"],
  theme: {
    colors: {
      'teal': '#226d87',
      'blue': {
        DEFAULT: '#021c51',
        'light-blue': '#ced9e1',
        'dark-blue': '#0a192f',
        500: '#2196f3'
      },
      'blue-dark': '#021c51',
      'blue-light': '#ced9e1',
      'blue-gray': '#22262a',
      'blue-500': '#2196f3',
      'pink': '#be5bda',
      'orange': '#ff7849',
      'green': {
        DEFAULT: '#58837e',
        'light-green': '#b5e7a0',
        100: '#dcfce7',
        600: '#16a34a'
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
    fontFamily: {
      sans: ['Montserrat', 'sans-serif'],
      serif: ['Georgia', 'serif'],
    },
    extend: {},
  },

  plugins: [
      require('@tailwindcss/forms'),
    ]
}
