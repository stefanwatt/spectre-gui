import { writable } from "svelte/store";

/**@type {import("svelte/store").Writable<RipgrepMatch>}*/
export const selected_match = writable(null);


/**@type {import("svelte/store").Writable<RipgrepResult[]>}*/
export const results = writable([]);
