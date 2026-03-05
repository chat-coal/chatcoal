/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts}'],
  theme: {
    extend: {
      colors: {
        ember: {
          DEFAULT: '#E8521A',
          dim: '#D4782A',
          glow: '#FF6B35',
          deep: '#8B2500',
        },
        coal: {
          950: '#070707',
          900: '#0D0D0D',
          800: '#141414',
          700: '#1C1C1C',
          600: '#252525',
          500: '#333333',
        },
      },
      fontFamily: {
        display: ['Syne', 'sans-serif'],
        body: ['Manrope', 'sans-serif'],
      },
      backgroundImage: {
        'ember-radial': 'radial-gradient(ellipse at center, #E8521A22 0%, transparent 70%)',
        'ember-cone': 'radial-gradient(ellipse at 50% 100%, #E8521A33 0%, transparent 60%)',
      },
      animation: {
        'ember-pulse': 'emberPulse 3s ease-in-out infinite',
        'float-up': 'floatUp 6s ease-in-out infinite',
        'glow-flicker': 'glowFlicker 4s ease-in-out infinite',
        'slide-up': 'slideUp 0.6s ease-out forwards',
        'fade-in': 'fadeIn 0.8s ease-out forwards',
      },
      keyframes: {
        emberPulse: {
          '0%, 100%': { opacity: '0.6', transform: 'scale(1)' },
          '50%': { opacity: '1', transform: 'scale(1.05)' },
        },
        floatUp: {
          '0%': { transform: 'translateY(0px)', opacity: '0.7' },
          '50%': { transform: 'translateY(-12px)', opacity: '1' },
          '100%': { transform: 'translateY(0px)', opacity: '0.7' },
        },
        glowFlicker: {
          '0%, 100%': { boxShadow: '0 0 20px #E8521A44' },
          '33%': { boxShadow: '0 0 40px #E8521A88' },
          '66%': { boxShadow: '0 0 25px #E8521A55' },
        },
        slideUp: {
          from: { opacity: '0', transform: 'translateY(30px)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
        fadeIn: {
          from: { opacity: '0' },
          to: { opacity: '1' },
        },
      },
    },
  },
  plugins: [],
}
