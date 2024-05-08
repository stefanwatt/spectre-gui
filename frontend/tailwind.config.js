import daisyui from "daisyui";
import catppuccin from "@catppuccin/tailwindcss";

/** @type {import('tailwindcss').Config}*/
const config = {
  content: ["./src/**/*.{html,js,svelte,ts}"],

  theme: {
    extend: {},
  },

  plugins: [
    daisyui,
    catppuccin({
      defaultFlavour: 'frappe',
    })
  ],
  daisyui: {
    themes: [
      {
        "catppuccin-frappe": {
          primary: "#8caaee",
          secondary: "#f4b8e4",
          accent: "#81c8be",
          neutral: "#232634",
          "base-100": "#303446",
          info: "#85c1dc",
          success: "#a6d189",
          warning: "#e5c890",
          error: "#e78284",
        },
      },
    ],
  },
};

module.exports = config;
