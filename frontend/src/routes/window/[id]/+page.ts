import type { PageLoad } from './$types';
import { App as wails } from '@bindings/nvim-gui/index.js';
import { error } from '@sveltejs/kit';

export const prerender = false;

export const load: PageLoad = async ({ params, url }) => {
	const gridIdParam = url.searchParams.get('gridId');
	const gridId = gridIdParam ? parseInt(gridIdParam) : null;
	const res = await wails.GetWindow(+params.id!);
	if (!res) {
		error(404, `could not find window with id=${+params.id}`);
	}
	const windowApi = {
		...res,
		mode: res.mode as App.VimMode,
		floatingWindows: [] as App.FloatingWindow[],
		cursor: res.cursor as { row: number; col: number }
	};
	if (!windowApi.cursor || !windowApi.id) {
		error(500, 'window object malformed');
	}

	const _window = windowApi satisfies App.NvimWindow;
	return { window: _window, gridId };
};
