import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
	plugins: [vue()],
	// Der Produktions-Build landet dort, wo ihn das Go-Binary einbettet
	// (internal/web/dist) — ein Artefakt, ein Deploy.
	build: {
		outDir: '../internal/web/dist',
		emptyOutDir: true
	},
	server: {
		port: 5174,
		proxy: {
			'/api': 'http://localhost:8080'
		}
	}
});
