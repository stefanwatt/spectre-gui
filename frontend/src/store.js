import { writable } from "svelte/store";

/**@type {import("svelte/store").Writable<RipgrepMatch>}*/
export const selectedProject = writable(null);
