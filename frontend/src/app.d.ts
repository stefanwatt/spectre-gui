import type { SvelteComponent } from "svelte";
import type { Writable as _Writable } from "svelte/store";
// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
declare global {
  namespace App {
    type NotificationLevel = "info" | "success" | "warning" | "error"

    interface SearchFlag {
      icon: SvelteComponent
      text: "match_whole_word" | "case_sensitive" | "regex"
    }

    interface Toast {
      level: NotificationLevel;
      text: string;
    }

    type Modifier = 'c' | 's' | 'a'

    interface RipgrepResult {
      Path: string;
      Matches: RipgrepMatch[]
    }

    type VimMode = "n" | "i" | "v" | "V" | "c"

    interface HighlightToken {
      text: string;
      start_row: number;
      end_row: number;
      start_col: number;
      end_col: number;
      foreground: string;
      background: string
      reverse: boolean;
      underline: boolean;
      undercurl: boolean;
      strikethrough: boolean;
      bold: boolean
      italic: boolean;
    }

    interface BufLine {
      sign: string;
      row: number;
      tokens: HighlightToken[]
    }

    interface CursorMoveEvent {
      row: number;
      col: number;
      key: string;
      top_line: number;
      bottom_line: number;
    }

    interface RipgrepMatch {
      Id: string;
      FileName: string;
      AbsolutePath: string;
      MatchedLine: string;
      TextBeforeMatch: string;
      TextAfterMatch: string;
      MatchedText: string;
      ReplacementText: string;
      Row: number;
      Col: number;
      Html: string;
    }

    type Writable<T> = _Writable<T>

    interface Keymap {
      mods: Modifier[];
      key: string;
      action: (e: KeyboardEvent) => void;
    }

    interface SearchResult {
      GroupedMatches: RipgrepResult[]
      PageIndex: number
      TotalPages: number
      TotalResults: number
      TotalFiles: number
    }

    interface ToastEvent {
      level: NotificationLevel;
      message: string;
    }
    interface State {
      SearchTerm: string
      ReplaceTerm: string
      Dir: string
      Include: string
      Exclude: string
      CaseSensitive: string
      Regex: string
      MatchWholeWord: string
      PreserveCase: string
    }
  }
}


export { };
