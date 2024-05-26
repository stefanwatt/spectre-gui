import { show_toast } from '$lib/notification/notification.service.js';
import { search } from '$lib/results/results.service';
import { EventsOn } from '$lib/wailsjs/runtime/runtime.js';
import { goto } from '$app/navigation'
import {
  search_term,
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
const CHANGE_URL = "change-url"

export function listen_for_events() {
  EventsOn(CHANGE_URL, (page) => {
    alert(`goto page ${page}`)
    goto(page)
  });
  EventsOn(REPLACE, () => {
    search(
      get(search_term),
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
