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

    interface RipgrepMatch {
      Id: string;
      Path: string;
      Row: number;
      Col: number;
      MatchedLine: string;
      MatchedText: string;
      TextBeforeMatch: string;
      TextAfterMatch: string;
      ReplacementText: string;
      Html: string;
    }
    type Writable<T> = _Writable<T>

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
