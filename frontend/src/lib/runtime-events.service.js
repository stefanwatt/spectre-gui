import { show_toast } from '$lib/notification/notification.service.js';
import { search } from '$lib/results/results.service';
import { EventsOn } from '$lib/wailsjs/runtime/runtime.js';
import {
  search_term,
  replace_term,
  case_sensitive,
  regex,
  match_whole_word,
  preserve_case,
  dir,
  include,
  exclude,
} from '$lib/store';
import { get } from 'svelte/store';



// const DELETE = "file-deleted"
const REPLACE = "file-replaced"
const REPLACE_ALL = "replaced-all"
const UNDO = "undo"
const TOAST = "toast"

export function listen_for_events() {
  EventsOn(REPLACE, () => {
    search(
      get(search_term),
      get(replace_term),
      get(dir),
      get(exclude),
      get(include),
      get(case_sensitive),
      get(regex),
      get(match_whole_word),
      get(preserve_case)
    );
  });

  EventsOn(REPLACE_ALL, () => {
    search(
      get(search_term),
      get(replace_term),
      get(dir),
      get(exclude),
      get(include),
      get(case_sensitive),
      get(regex),
      get(match_whole_word),
      get(preserve_case)
    );
  });

  EventsOn(UNDO, () => {
    search(
      get(search_term),
      get(replace_term),
      get(dir),
      get(exclude),
      get(include),
      get(case_sensitive),
      get(regex),
      get(match_whole_word),
      get(preserve_case)
    );
  });

  EventsOn(TOAST, /** 
    @param {App.NotificationLevel} level 
    @param {string} message 
  */(level, message) => {
      show_toast(level, message);
    });
}
