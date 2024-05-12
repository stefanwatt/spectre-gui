import { writable } from "svelte/store";

/**@type {import("svelte/store").Writable<App.RipgrepMatch|null>}*/
export const selected_match = writable(null);
/**@type {import("svelte/store").Writable<App.RipgrepResult[]>}*/
export const results = writable([]);
/**@type {import("svelte/store").Writable<App.Toast|null>}*/
export const toast = writable();
/**@type {import("svelte/store").Writable<string>}*/
export const search_term = writable("");
/**@type {import("svelte/store").Writable<string>}*/
export const replace_term = writable("");
/**@type {import("svelte/store").Writable<string>}*/
export const dir = writable("");
/**@type {import("svelte/store").Writable<string>}*/
export const include = writable("");
/**@type {import("svelte/store").Writable<string>}*/
export const exclude = writable("");

/**@type {import("svelte/store").Writable<boolean>}*/
export const case_sensitive = writable(false);
/**@type {import("svelte/store").Writable<boolean>}*/
export const regex = writable(false);
/**@type {import("svelte/store").Writable<boolean>}*/
export const match_whole_word = writable(false);
/**@type {import("svelte/store").Writable<boolean>}*/
export const preserve_case = writable(true);
/**@type {import("svelte/store").Writable<App.SearchFlag[]>}*/
