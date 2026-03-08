import type { PageLoad } from './$types';
import { App as wails } from '@bindings/nvim-gui/index.js';
import { error } from '@sveltejs/kit';

export const prerender = false;

export const load: PageLoad = async ({ params, url }) => {
	const gridIdParam = url.searchParams.get('gridId');
	const gridId = gridIdParam ? parseInt(gridIdParam) : null;
	const winId = +params.id!;
	const res = await wails.GetWindow(winId);
	if (!res) {
		error(404, `could not find float window with id=${winId}`);
	}
	return { winId, gridId };
};
