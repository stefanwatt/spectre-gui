import type { SvelteComponent } from "svelte";
import type { Writable as _Writable } from "svelte/store";
// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
declare global {

  namespace App {
    interface NvimPosition {
      row: number;
      col: number;
    }

    interface NvimCell {
      char: string;
      fg: string;
      bg: string;
      classes: string;
      highlight: string;
    }

    interface NvimHighlight {
      id: string;
      [key: string]: any;
    }
    type VimMode = "normal" | "insert" | "visual" | "cmdline_normal"


    interface CursorMoveEvent {
      row: number;
      col: number;
      key: string;
      top_line: number;
      bottom_line: number;
    }

    type Writable<T> = _Writable<T>

    interface Keymap {
      mods: Modifier[];
      key: string;
      action: (e: KeyboardEvent) => void;
    }
  }
}


export { };
