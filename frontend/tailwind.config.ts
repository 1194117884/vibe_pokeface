import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        "starbucks": "#006241",
        "green-accent": "#00754A",
        "house-green": "#1E3932",
        "green-uplift": "#2b5148",
        "green-light": "#d4e9e2",
        "gold": "#cba258",
        "gold-light": "#dfc49d",
        "gold-lightest": "#faf6ee",
        "cream": "#f2f0eb",
        "ceramic": "#edebe9",
        "neutral-cool": "#f9f9f9",
        "text-black": "rgba(0, 0, 0, 0.87)",
        "text-black-soft": "rgba(0, 0, 0, 0.58)",
        "text-white-soft": "rgba(255, 255, 255, 0.70)",
        "red-error": "#c82014",
        "yellow-warn": "#fbbc05",
      },
      fontFamily: {
        sans: ['Inter', '"Helvetica Neue"', 'Helvetica', 'Arial', 'sans-serif'],
      },
      borderRadius: {
        pill: "50px",
      },
      spacing: {
        "space-1": "0.4rem",
        "space-2": "0.8rem",
        "space-3": "1.6rem",
        "space-4": "2.4rem",
        "space-5": "3.2rem",
        "space-6": "4rem",
        "space-7": "4.8rem",
        "space-8": "5.6rem",
        "space-9": "6.4rem",
      },
      boxShadow: {
        card: "0 0 0.5px rgba(0,0,0,0.14), 0 1px 1px rgba(0,0,0,0.24)",
        nav: "0 1px 3px rgba(0,0,0,0.1), 0 2px 2px rgba(0,0,0,0.06), 0 0 2px rgba(0,0,0,0.07)",
        frap: "0 0 6px rgba(0,0,0,0.24), 0 8px 12px rgba(0,0,0,0.14)",
      },
      letterSpacing: {
        tight: "-0.01em",
      },
    },
  },
  plugins: [],
};

export default config;
