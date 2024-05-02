/** @type {import('tailwindcss').Config}*/
const config = {
  content: ["./src/**/*.{html,js,svelte,ts}"],

  theme: {
    extend: {},
  },

  plugins: [require("@catppuccin/tailwindcss")({
    defaultFlavour: 'frappe',
  })]
};

module.exports = config;
