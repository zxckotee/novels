/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Основные цвета (темная тема Remanga-style)
        background: {
          primary: '#0a0a0a',
          secondary: '#121212',
          tertiary: '#1a1a1a',
          card: '#1e1e1e',
          hover: '#252525',
        },
        foreground: {
          primary: '#ffffff',
          secondary: '#b3b3b3',
          tertiary: '#808080',
          muted: '#666666',
        },
        // Акцентные цвета
        accent: {
          primary: '#8b5cf6',     // Фиолетовый
          secondary: '#f97316',   // Оранжевый
          success: '#22c55e',     // Зеленый
          warning: '#eab308',     // Желтый
          danger: '#ef4444',      // Красный
          info: '#3b82f6',        // Синий
        },
        // Бордеры
        border: {
          primary: '#2a2a2a',
          secondary: '#333333',
          focus: '#8b5cf6',
        },
      },
      borderRadius: {
        'card': '12px',
        'button': '8px',
        'input': '8px',
        'tag': '6px',
      },
      fontFamily: {
        sans: ['Inter', 'Roboto', 'system-ui', 'sans-serif'],
        heading: ['Manrope', 'Inter', 'sans-serif'],
      },
      fontSize: {
        'xs': ['0.75rem', { lineHeight: '1rem' }],
        'sm': ['0.875rem', { lineHeight: '1.25rem' }],
        'base': ['1rem', { lineHeight: '1.5rem' }],
        'lg': ['1.125rem', { lineHeight: '1.75rem' }],
        'xl': ['1.25rem', { lineHeight: '1.75rem' }],
        '2xl': ['1.5rem', { lineHeight: '2rem' }],
        '3xl': ['1.875rem', { lineHeight: '2.25rem' }],
        '4xl': ['2.25rem', { lineHeight: '2.5rem' }],
      },
      spacing: {
        '18': '4.5rem',
        '22': '5.5rem',
        '30': '7.5rem',
      },
      boxShadow: {
        'card': '0 4px 6px -1px rgba(0, 0, 0, 0.3), 0 2px 4px -2px rgba(0, 0, 0, 0.2)',
        'card-hover': '0 10px 15px -3px rgba(0, 0, 0, 0.4), 0 4px 6px -4px rgba(0, 0, 0, 0.3)',
        'dropdown': '0 10px 25px -5px rgba(0, 0, 0, 0.5)',
      },
      animation: {
        'fade-in': 'fadeIn 0.2s ease-in-out',
        'slide-up': 'slideUp 0.2s ease-out',
        'slide-down': 'slideDown 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
      },
      aspectRatio: {
        'cover': '2 / 3',
        'banner': '16 / 9',
        'hero': '3 / 2',
      },
      gridTemplateColumns: {
        'novels': 'repeat(auto-fill, minmax(150px, 1fr))',
        'novels-lg': 'repeat(auto-fill, minmax(180px, 1fr))',
      },
    },
  },
  plugins: [
    // Плагин для скроллбара
    function({ addUtilities }) {
      addUtilities({
        '.scrollbar-hide': {
          '-ms-overflow-style': 'none',
          'scrollbar-width': 'none',
          '&::-webkit-scrollbar': {
            display: 'none',
          },
        },
        '.scrollbar-thin': {
          'scrollbar-width': 'thin',
          '&::-webkit-scrollbar': {
            width: '6px',
            height: '6px',
          },
          '&::-webkit-scrollbar-track': {
            background: '#1a1a1a',
          },
          '&::-webkit-scrollbar-thumb': {
            background: '#333333',
            borderRadius: '3px',
          },
          '&::-webkit-scrollbar-thumb:hover': {
            background: '#444444',
          },
        },
      });
    },
  ],
};
