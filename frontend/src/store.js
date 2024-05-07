import { writable } from "svelte/store";

/**@type {import("svelte/store").Writable<RipgrepMatch>}*/
export const selected_match = writable(null);
/**@type {import("svelte/store").Writable<RipgrepResult[]>}*/
export const results = writable([]);
/**@type {import("svelte/store").Writable<Toast>}*/
export const toast = writable();
/**@type {import("svelte/store").Writable<string>}*/
export const search_term = writable("utils");
/**@type {import("svelte/store").Writable<string>}*/
export const replace_term = writable("foo");
/**@type {import("svelte/store").Writable<string>}*/
export const dir = writable("/home/stefan/Projects/bubbletube");
/**@type {import("svelte/store").Writable<string>}*/
export const include = writable("**/*.go");
/**@type {import("svelte/store").Writable<string>}*/
export const exclude = writable("**/*.sh");
