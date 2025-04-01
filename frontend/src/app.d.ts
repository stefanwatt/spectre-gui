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

    interface NvimWindow {
      id: number;
      content: NvimCell[][];
      type: string;
      width: number;
      height: number;
    }

    interface CmdLine {
      visible: boolean;
      firstc?: string;
      prompt?: string;
      content?: string;
      indent?: number;
      pos?: number;
    }

    interface NvimHighlight {
      id: number;
      fg: string;
      bg: string;
      bold: boolean;
      italic: boolean;
      underline: boolean;
      undercurl: boolean;
      strikethrough: boolean;
      reverse: boolean;
    }
    type VimMode = "normal" | "insert" | "visual" | "cmdline_normal" | "cmdline_insert"
    interface GridProps {
      content: App.NvimCell[][]
      decode?: (input: string) => string
    }

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

    interface FloatingWindow {
      id: number;
      grid_id: number;
      anchor_grid: number;
      anchor: string;
      row: number;
      col: number;
      width: number;
      height: number;
      z_index: number;
      focusable: boolean;
      is_popup: boolean;
      grid: NvimCell[][];
    }
  }
}


export { };
