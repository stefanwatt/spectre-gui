import type { SvelteComponent } from "svelte";
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
    }

    interface ToastEvent {
      level: NotificationLevel;
      message: string;
    }

  }
}


export { };
