import MatchWholeWord from "./icons/MatchWholeWord.svelte"
import CaseSensitive from "./icons/CaseSensitive.svelte"
import Regex from "./icons/Regex.svelte"

type NotificationLevel = "info" | "success" | "warning" | "error"
interface SearchFlag {
  icon: MatchWholeWord | CaseSensitive | Regex
  value: "match_whole_word" | "case_sensitive" | "regex"
}

interface Toast {
  level: NotificationLevel;
  text: string;
}

type Modifier = 'c' | 's' | 'a'

interface RipgrepResult {
  path: string;
  matches: RipgrepMatch[]
}

interface RipgrepMatch {
  Id: string;
  Path: string;
  Row: number;
  Col: number;
  MatchedLine: string;

}

type RipgrepResultApi = { [key: string]: Array<RipgrepMatch> };
