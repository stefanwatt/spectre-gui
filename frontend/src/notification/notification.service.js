import { toast as toast$ } from "../store";

/**
 * @param {NotificationLevel} level
 * @param {string} text
 * */
export function show_toast(level, text) {
  toast$.set({ level, text })
  setTimeout(() => {
    toast$.set(null)
  }, 2000);
}
