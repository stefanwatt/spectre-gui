import { writable } from "svelte/store";

/**@type {App.Writable<App.RipgrepMatch|null>}*/
export const selected_match = writable(null);
/**@type {App.Writable<App.RipgrepResult[]>}*/
export const results = writable([]);
/**@type {App.Writable<App.Toast|null>}*/
export const toast = writable();
/**@type {App.Writable<string>}*/
export const search_term = writable("");
/**@type {App.Writable<string>}*/
export const replace_term = writable("");
/**@type {App.Writable<string>}*/
export const dir = writable("");
/**@type {App.Writable<string>}*/
export const include = writable("");
/**@type {App.Writable<string>}*/
export const exclude = writable("");
/**@type {App.Writable<boolean>}*/
export const case_sensitive = writable(false);
/**@type {App.Writable<boolean>}*/
export const regex = writable(false);
/**@type {App.Writable<boolean>}*/
export const match_whole_word = writable(false);
/**@type {App.Writable<boolean>}*/
export const preserve_case = writable(true);
/**@type {App.Writable<number>}*/
export const total_pages = writable(0);
/**@type {App.Writable<number>}*/
export const page_index = writable(0);
/**@type {App.Writable<number>}*/
export const total_results = writable(0);
/**@type {App.Writable<number>}*/
export const total_files = writable(0);

