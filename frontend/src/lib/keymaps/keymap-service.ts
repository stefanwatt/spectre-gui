import { Clipboard } from '@wailsio/runtime';
import { Call } from '@wailsio/runtime';
import { getKeymapMode } from '$lib/state.svelte';

let activeKeymaps = new Map<string, App.Keymap>();
const pickerModes: App.KeymapMode[] = [
	'find-references',
	'find-files',
	'live-grep',
	'find-buffer-symbols'
];

export function registerKeymap(keymap: App.Keymap) {
	const keymapString = keymapToString(keymap);
	if (activeKeymaps.has(keymapString)) {
		const handler = activeKeymaps.get(keymapString)!.action;
		window.removeEventListener('keydown', handler);
		activeKeymaps.delete(keymapString);
	}
	activeKeymaps.set(keymapString, keymap);
}

export function handleKeypress(event: KeyboardEvent) {
	const keymapMode = getKeymapMode();
	const keymapString = eventToKeymapString(event);
	if (activeKeymaps.has(keymapString)) {
		event.preventDefault();
		activeKeymaps.get(keymapString)!.action(event);
		return;
	}
	if (pickerModes.includes(keymapMode)) return;
	event.preventDefault();
	if (event.ctrlKey && event.shiftKey && event.key === 'V') {
		Clipboard.Text().then((text: string) =>
			Call.ByName('main.App.Paste', text).catch((err) => console.error('Paste failed', err))
		);
		return;
	}
	Call.ByName('main.App.SendKey', event.key, event.ctrlKey, event.altKey, event.shiftKey, keymapMode).catch(
		(err) => console.error('SendKey failed', err)
	);
}

function eventToKeymapString(event: KeyboardEvent) {
	const mods = [];
	if (event.altKey) mods.push('a');
	if (event.ctrlKey) mods.push('c');
	if (event.shiftKey) mods.push('s');
	if (!mods.length) return event.key + '-' + getKeymapMode();
	return mods.sort().join('-') + '-' + event.key + '-' + getKeymapMode();
}

function keymapToString(keymap: App.Keymap) {
	if (!keymap.mods?.length) return keymap.key + '-' + keymap.mode;
	return keymap.mods.sort().join('-') + '-' + keymap.key + '-' + keymap.mode;
}

function modsPressed(mods: App.Modifier[], event: KeyboardEvent): boolean {
	if (!mods.length) return true;
	return mods.every((mod) => {
		switch (mod) {
			case 'c':
				return event.ctrlKey;
			case 's':
				return event.shiftKey;
			case 'a':
				return event.altKey;
			default:
				break;
		}
	});
}
