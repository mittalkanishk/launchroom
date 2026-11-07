// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://starlight.astro.build/reference/configuration/
export default defineConfig({
	integrations: [
		starlight({
			title: 'Your docs',
			sidebar: [
				{
					label: 'Guides',
					// Each item here is one entry in the navigation menu.
					items: [{ label: 'Getting started', slug: 'guides/getting-started' }],
				},
				{
					label: 'Reference',
					// Every page in src/content/docs/reference is listed here.
					items: [{ autogenerate: { directory: 'reference' } }],
				},
			],
		}),
	],
});
