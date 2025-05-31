// @ts-check
import { defineConfig } from 'astro/config';
import node from '@astrojs/node';
import tailwindcss from '@tailwindcss/vite';

// https://astro.build/config
export default defineConfig({
  output: 'static',
  adapter: node({
    mode: 'standalone'
  }),
  image: {
    service: {
      entrypoint: "astro/assets/services/sharp",
      config: {},
    },
  },
  vite: {
    plugins: [tailwindcss()]
  },
  server: {
    port: 4321,
    host: true
  }
});
