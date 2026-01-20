/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./examples/**/*.{templ,go,html}",
    "./pkg/**/*.{templ,go,html}",
    "./static/**/*.html",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
