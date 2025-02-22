import {writable }from "svelte/store"

export const mode = writable<App.VimMode>('n')
