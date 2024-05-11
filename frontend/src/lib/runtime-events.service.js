import { show_toast } from '$lib/notification/notification.service.js';
import { search } from '$lib/results/results.service';
import { EventsOn } from '$lib/wailsjs/runtime/runtime.js';
import { search_flags, search_term, replace_term, preserve_case, dir, include, exclude, } from '$lib/store';
import { get } from 'svelte/store';



// const DELETE = "file-deleted"
const REPLACE = "file-replaced"
const REPLACE_ALL = "replaced-all"
const UNDO = "undo"
const TOAST = "toast"

export function listen_for_events() {
  EventsOn(REPLACE, () => {
    const flags = get(search_flags).map((f) => f.text);
    console.log('searching with flags ', flags);
    search(get(search_term), get(dir), get(include), get(exclude), flags, get(replace_term), get(preserve_case));
  });

  EventsOn(REPLACE_ALL, () => {
    const flags = get(search_flags).map((f) => f.text);
    console.log('searching with flags ', flags);
    search(get(search_term), get(dir), get(include), get(exclude), flags, get(replace_term), get(preserve_case));
  });

  EventsOn(UNDO, () => {
    const flags = get(search_flags).map((f) => f.text);
    console.log('searching with flags ', flags);
    search(get(search_term), get(dir), get(include), get(exclude), flags, get(replace_term), get(preserve_case));
  });

  EventsOn(TOAST, /** 
    @param {App.NotificationLevel} level 
    @param {string} message 
  */(level, message) => {
      show_toast(level, message);
    });
}
