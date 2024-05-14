export const proxy = {
	"/dev": {
		target: "http://127.0.0.1:8001",
		changeOrigin: true,
		rewrite: (path: string) => path.replace(/^\/dev/, "")
	},
	"/v2": {
		target: "http://127.0.0.1:8001",
		changeOrigin: true
	},
	"/fc": {
		target: "http://127.0.0.1:8001",
		changeOrigin: true
	},
	"/cdn": {
		target: "http://127.0.0.1:8001",
		changeOrigin: true
	},
	"/prod": {
		target: "https://show.cool-admin.com",
		changeOrigin: true,
		rewrite: (path: string) => path.replace(/^\/prod/, "/api")
	}
};
