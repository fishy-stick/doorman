import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const backendTarget = process.env.DOORMAN_BACKEND_URL ?? 'http://127.0.0.1:8080'

export default defineConfig(({ command, mode }) => {
  const isDevServer = command === 'serve'
  const isEmbedBuild = mode === 'embed'

  return {
    base: isDevServer || isEmbedBuild ? '/admin/' : './',
    plugins: [react()],
    build: {
      outDir: isEmbedBuild ? '../internal/webui/dist' : 'dist',
      emptyOutDir: true,
    },
    server: {
      proxy: {
        '/admin/api': {
          target: backendTarget,
          changeOrigin: true,
        },
        '/knock': {
          target: backendTarget,
          changeOrigin: true,
        },
      },
    },
  }
})
