import { Call } from '@wailsio/runtime';

export const keymaps: App.Keymap[] = [
  {
    key: 'Tab',
    mode: 'cmdline',
    mods: [],
		action: (e) => {
			e.preventDefault()
			Call.ByName('main.App.SubstituteJump').catch((err) => {
				console.error('SubstituteJump failed', err);
			})
		}
	},
]
