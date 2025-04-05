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

    interface NvimToken {
      text: string;
      fg: string;
      bg: string;
      classes: string;
      highlight: string;
    }

    interface NvimRow{
      index:number;
      tokens: NvimToken[]
    }

    interface NvimWindowMap {
      [key: number]: NvimRow[];
    }
    interface NvimContent {
      winId: number
      content: NvimToken[][]
    }

    interface NvimLayout {
      cols: string
      rows: string
      activeWindowId: number
      windows: NvimWindow[]
    }

    interface NvimWindow {
      id: number;
      type: string;
      width: number;
      height: number;
      colStart: number;
      colEnd: number;
      rowStart: number;
      rowEnd: number;
      filetype: string;
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
      content?: App.NvimRow[]
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
      gridId: number;
      anchorWindow: number;
      anchor: string;
      row: number;
      col: number;
      width: number;
      height: number;
      zIndex: number;
      focusable: boolean;
      isPopup: boolean;
      isHex: boolean;
      filetype:string
    }
  }
}


export { };
