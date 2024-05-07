type NotificationLevel = "info" | "success" | "warning" | "error"

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
